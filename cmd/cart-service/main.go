package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
	cartv1 "zjMall/gen/go/api/proto/cart"
	commonv1 "zjMall/gen/go/api/proto/common"
	"zjMall/internal/cart-service/handler"
	"zjMall/internal/cart-service/repository"
	"zjMall/internal/cart-service/service"
	"zjMall/internal/common/cache"
	"zjMall/internal/common/client"
	"zjMall/internal/common/middleware"
	"zjMall/internal/common/mq"
	registry "zjMall/internal/common/register"
	"zjMall/internal/common/server"
	"zjMall/internal/config"
	"zjMall/internal/database"
	"zjMall/pkg"

	"google.golang.org/grpc"
)

const serviceName = "cart-service"
const serviceIP = "127.0.0.1"

func main() {
	// 0. 初始化日志：同时输出到控制台和文件 logs/cart-service.log
	logDir := fmt.Sprintf("./logs/%s", serviceName)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("Error creating log directory: %v", err)
	}
	logFilePath := filepath.Join(logDir, serviceName+time.Now().Format("20060102150405")+".log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
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
	//初始化JWT
	pkg.InitJWT(cfg.GetJWTConfig())
	// 2. 初始化数据库（购物车数据存储在 MySQL）
	mysqlConfig, err := cfg.GetDatabaseConfigForService(serviceName)
	if err != nil {
		log.Fatalf("Error getting database config for %s: %v", serviceName, err)
	}
	db, err := database.InitMySQL(mysqlConfig)
	if err != nil {
		log.Fatalf("Error initializing MySQL: %v", err)
	}
	defer database.CloseMySQL()

	// 3. 初始化 Redis（用于缓存）
	redisConfig := cfg.GetRedisConfig()
	redisClient, err := database.InitRedis(redisConfig)
	if err != nil {
		log.Fatalf("Error initializing Redis: %v", err)
	}
	defer database.CloseRedis()

	// 4. 创建通用的缓存仓库
	baseCacheRepo := cache.NewCacheRepository(redisClient)

	// 5. 初始化 RabbitMQ（可选，如果配置了才初始化）
	var mqProducer mq.MessageProducer
	rabbitCfg := cfg.GetRabbitMQConfig()
	if rabbitCfg != nil && rabbitCfg.Host != "" {
		ch, err := database.InitRabbitMQ(rabbitCfg)
		if err != nil {
			log.Printf("⚠️ RabbitMQ 初始化失败，将使用同步模式: %v", err)
		} else {
			defer database.CloseRabbitMQ()
			mqProducer = mq.NewMessageProducer(ch, rabbitCfg.Queue)
			log.Printf("✅ RabbitMQ 初始化成功，队列=%s", rabbitCfg.Queue)

			// 启动购物车事件消费者：从 MQ 同步数据到 MySQL
			consumerCtx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go mq.StartCartEventConsumer(consumerCtx, db, ch, rabbitCfg.Queue)
		}
	} else {
		log.Println("ℹ️ 未配置 RabbitMQ，将使用同步模式（Redis + MySQL 双写）")
	}

	// 6. 创建购物车仓库（Redis 主存储 + MQ 异步同步到 MySQL）
	log.Printf("🔍 [DEBUG] 创建 CartRepository，mqProducer 是否为 nil: %v", mqProducer == nil)
	cartRepo := repository.NewCartRepository(db, redisClient, baseCacheRepo, mqProducer)

	// 7. 初始化商品服务客户端（优先通过 Nacos 发现，其次使用配置中的备用地址）
	var productClient client.ProductClient
	productServiceAddr := ""

	// 7.1 尝试从 Nacos 发现 product-service
	productServiceAddr, err = registry.SelectOneHealthyInstance(nacosClient, "product-service")
	if err != nil {
		log.Printf("⚠️ 从 Nacos 发现商品服务失败，将尝试使用配置中的备用地址: %v", err)
	}

	// 7.2 如果 Nacos 没有可用实例，则回退到配置文件中的地址
	if productServiceAddr == "" {
		serviceClientsConfig := cfg.GetServiceClientsConfig()
		if serviceClientsConfig.ProductServiceAddr != "" {
			productServiceAddr = serviceClientsConfig.ProductServiceAddr
			log.Printf("ℹ️ 使用配置中的商品服务备用地址: %s", productServiceAddr)
		}
	}

	// 7.3 如果拿到了地址，则创建 gRPC 客户端
	if productServiceAddr != "" {
		productClient, err = client.NewProductClient(productServiceAddr)
		if err != nil {
			log.Printf("⚠️ 商品服务客户端初始化失败，购物车功能可能受限: %v", err)
		} else {
			defer productClient.Close()
			log.Printf("✅ 商品服务客户端连接成功: %s", productServiceAddr)
		}
	} else {
		log.Println("ℹ️ 未找到商品服务地址，将使用模拟数据")
	}

	// 8. 创建购物车服务
	cartService := service.NewCartService(cartRepo, productClient)

	// 9. 创建购物车 Handler
	cartServiceHandler := handler.NewCartServiceHandler(cartService)

	// 10. 获取服务配置
	serviceCfg, err := cfg.GetServiceConfig(serviceName)
	if err != nil {
		log.Fatalf("Error getting service config: %v", err)
	}

	// 11. 创建服务器实例
	srv := server.NewServer(&server.Config{
		GRPCAddr: fmt.Sprintf(":%d", serviceCfg.GRPC.Port),
		HTTPAddr: fmt.Sprintf(":%d", serviceCfg.HTTP.Port),
	})

	// 12. 注册 gRPC 服务
	srv.RegisterGRPCService(func(grpcServer *grpc.Server) {
		cartv1.RegisterCartServiceServer(grpcServer, cartServiceHandler)
	})

	// 13. 注册 HTTP 网关处理器
	if err := srv.RegisterHTTPGateway(commonv1.RegisterHealthServiceHandlerFromEndpoint); err != nil {
		log.Fatalf("failed to register health service gateway: %v", err)
	}

	if err := srv.RegisterHTTPGateway(cartv1.RegisterCartServiceHandlerFromEndpoint); err != nil {
		log.Fatalf("failed to register cart service gateway: %v", err)
	}

	// 14. 注册 Swagger 文档
	srv.RegisterSwagger(
		server.SwaggerDoc{
			Name:        "cart",
			FilePath:    "docs/openapi/cart.swagger.json",
			Title:       "购物车服务 API",
			Description: "购物车服务 API 文档，包括添加商品、修改数量、删除商品等功能",
			Version:     "1.0.0",
		},
	)

	// 15. 注册中间件
	srv.UseMiddleware(
		middleware.CORS(middleware.DefaultCORSConfig()), // 1. 最外层：处理跨域
		middleware.Recovery(),                           // 2. 捕获 panic
		middleware.Logging(),                            // 3. 记录日志
		middleware.TraceID(),                            // 4. 生成 TraceID
		middleware.Auth(),                               // 5. 认证（购物车需要登录）
	)

	// 16. 启动服务器（阻塞）
	if err := srv.Start(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
