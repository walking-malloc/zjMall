package main

import (
	"fmt"
	"log"
	"math/rand"
	"path/filepath"
	"time"
	commonv1 "zjMall/gen/go/api/proto/common"
	orderv1 "zjMall/gen/go/api/proto/order"
	"zjMall/internal/common/client"
	"zjMall/internal/common/middleware"
	registry "zjMall/internal/common/register"
	"zjMall/internal/common/server"
	"zjMall/internal/config"
	"zjMall/internal/database"
	"zjMall/internal/order-service/handler"
	"zjMall/internal/order-service/repository"
	"zjMall/internal/order-service/service"
	"zjMall/pkg"

	"google.golang.org/grpc"
)

const serviceName = "order-service"
const serviceIP = "127.0.0.1"

func main() {
	logFile, err := pkg.InitLog(serviceName)
	if err != nil {
		log.Fatalf("Error initializing log: %v", err)
	}
	defer logFile.Close()
	log.Printf("==== %s starting ====", serviceName)

	// 1. 加载配置
	configPath := filepath.Join("./configs", "config.yaml")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	//2.初始化Nacos
	svcCfg, _ := cfg.GetServiceConfig(serviceName)
	nacosConfig := cfg.GetNacosConfig()
	nacosClient, err := registry.NewNacosNamingClient(nacosConfig)
	if err != nil {
		log.Fatalf("❌ Nacos 初始化失败: %v", err)
	}
	registry.RegisterService(nacosClient, serviceName, serviceIP, uint64(svcCfg.GRPC.Port))
	//初始化客户端
	productServiceAddr, err := registry.SelectOneHealthyInstance(nacosClient, "product-service")
	if err != nil || productServiceAddr == "" {
		log.Fatalf("❌ 从 Nacos 发现商品服务失败: %v", err)
	}
	productClient, err := client.NewProductClient(productServiceAddr)
	if err != nil {
		log.Fatalf("❌ 商品服务客户端初始化失败: %v", err)
	}
	defer productClient.Close()
	log.Printf("✅ 商品服务客户端连接成功: %s", productServiceAddr)

	inventoryServiceAddr, err := registry.SelectOneHealthyInstance(nacosClient, "inventory-service")
	if err != nil || inventoryServiceAddr == "" {
		log.Fatalf("❌ 从 Nacos 发现库存服务失败: %v", err)
	}
	inventoryClient, err := client.NewInventoryClient(inventoryServiceAddr)
	if err != nil {
		log.Fatalf("❌ 库存服务客户端初始化失败: %v", err)
	}
	defer inventoryClient.Close()
	log.Printf("✅ 库存服务客户端连接成功: %s", inventoryServiceAddr)

	userServiceAddr, err := registry.SelectOneHealthyInstance(nacosClient, "user-service")
	if err != nil || userServiceAddr == "" {
		log.Fatalf("❌ 从 Nacos 发现用户服务失败: %v", err)
	}
	userClient, err := client.NewUserClient(userServiceAddr)
	if err != nil {
		log.Fatalf("❌ 用户服务客户端初始化失败: %v", err)
	}
	defer userClient.Close()
	log.Printf("✅ 用户服务客户端连接成功: %s", userServiceAddr)

	cartServiceAddr, err := registry.SelectOneHealthyInstance(nacosClient, "cart-service")
	if err != nil || cartServiceAddr == "" {
		log.Fatalf("❌ 从 Nacos 发现购物车服务失败: %v", err)
	}
	cartClient, err := client.NewCartClient(cartServiceAddr)
	if err != nil {
		log.Fatalf("❌ 购物车服务客户端初始化失败: %v", err)
	}
	defer cartClient.Close()
	log.Printf("✅ 购物车服务客户端连接成功: %s", cartServiceAddr)
	// 初始化 JWT（如果订单需要鉴权）
	pkg.InitJWT(cfg.GetJWTConfig())

	//3.初始化redis
	redisConfig := cfg.GetRedisConfig()
	redisClient, err := database.InitRedis(redisConfig)
	if err != nil {
		log.Fatalf("Error initializing Redis: %v", err)
	}
	defer database.CloseRedis()

	// 2. 初始化数据库
	mysqlConfig, err := cfg.GetDatabaseConfigForService(serviceName)
	if err != nil {
		log.Fatalf("Error getting database config for %s: %v", serviceName, err)
	}
	db, err := database.InitMySQL(mysqlConfig)
	if err != nil {
		log.Fatalf("Error initializing MySQL: %v", err)
	}
	defer database.CloseMySQL()

	//初始化随机种子
	rand.New(rand.NewSource(time.Now().UnixNano()))
	// 3. 创建仓储与服务
	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, productClient, inventoryClient, userClient, cartClient, redisClient)
	orderHandler := handler.NewOrderServiceHandler(orderService)

	// 4. 获取服务配置
	serviceCfg, err := cfg.GetServiceConfig(serviceName)
	if err != nil {
		log.Fatalf("Error getting service config: %v", err)
	}

	// 5. 创建服务器实例
	srv := server.NewServer(&server.Config{
		GRPCAddr: fmt.Sprintf(":%d", serviceCfg.GRPC.Port),
		HTTPAddr: fmt.Sprintf(":%d", serviceCfg.HTTP.Port),
	})

	// 6. 注册 gRPC 服务
	srv.RegisterGRPCService(func(grpcServer *grpc.Server) {
		orderv1.RegisterOrderServiceServer(grpcServer, orderHandler)
	})

	// 7. 注册 HTTP 网关处理器
	if err := srv.RegisterHTTPGateway(commonv1.RegisterHealthServiceHandlerFromEndpoint); err != nil {
		log.Fatalf("failed to register health service gateway: %v", err)
	}
	if err := srv.RegisterHTTPGateway(orderv1.RegisterOrderServiceHandlerFromEndpoint); err != nil {
		log.Fatalf("failed to register order service gateway: %v", err)
	}

	srv.RegisterSwagger(
		server.SwaggerDoc{
			Name:        "order",
			FilePath:    "docs/openapi/order.swagger.json",
			Title:       "订单服务 API",
			Description: "订单服务 API 文档，包括订单创建、查询、取消等功能",
			Version:     "1.0.0",
		},
	)
	// 9. 注册中间件
	srv.UseMiddleware(
		middleware.CORS(middleware.DefaultCORSConfig()),
		middleware.Recovery(),
		middleware.Logging(),
		middleware.TraceID(),
		middleware.Auth(),
	)

	// 10. 启动服务器
	if err := srv.Start(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
