package main

import (
	"fmt"
	"log"
	"path/filepath"

	commonv1 "zjMall/gen/go/api/proto/common"
	inventoryv1 "zjMall/gen/go/api/proto/inventory"
	"zjMall/internal/common/authz"
	"zjMall/internal/common/middleware"
	registry "zjMall/internal/common/register"
	"zjMall/internal/common/server"
	"zjMall/internal/config"
	"zjMall/internal/database"
	invHandler "zjMall/internal/inventory-service/handler"
	"zjMall/internal/inventory-service/repository"
	"zjMall/internal/inventory-service/service"
	"zjMall/pkg"

	"google.golang.org/grpc"
)

const serviceName = "inventory-service"
const serviceIP = "127.0.0.1"

func main() {
	logFile, err := pkg.InitLog(serviceName)
	if err != nil {
		log.Fatalf("Error initializing log: %v", err)
	}
	defer logFile.Close()

	// 1. 加载配置
	configPath := filepath.Join("./configs", "config.yaml")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	// 加载完配置 cfg 之后：
	if err := authz.InitCasbin(); err != nil {
		log.Fatalf("❌ Casbin 初始化失败: %v", err)
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

	// // 3. 初始化 Redis（用于缓存）
	// redisConfig := cfg.GetRedisConfig()
	// redisClient, err := database.InitRedis(redisConfig)
	// if err != nil {
	// 	log.Fatalf("Error initializing Redis: %v", err)
	// }
	// defer database.CloseRedis()

	// 6. 创建购物车仓库（Redis 主存储 + MQ 异步同步到 MySQL）
	inventoryRepo := repository.NewStockRepository(db)

	// 9. 创建购物车服务
	inventoryService := service.NewInventoryService(inventoryRepo)

	// 10. 创建购物车 Handler
	inventoryServiceHandler := invHandler.NewInventoryHandler(inventoryService)

	// 11. 获取服务配置
	serviceCfg, err := cfg.GetServiceConfig(serviceName)
	if err != nil {
		log.Fatalf("Error getting service config: %v", err)
	}

	// 12. 创建服务器实例
	srv := server.NewServer(&server.Config{
		GRPCAddr: fmt.Sprintf(":%d", serviceCfg.GRPC.Port),
		HTTPAddr: fmt.Sprintf(":%d", serviceCfg.HTTP.Port),
	})

	// 13. 注册 gRPC 服务
	srv.RegisterGRPCService(func(grpcServer *grpc.Server) {
		inventoryv1.RegisterInventoryServiceServer(grpcServer, inventoryServiceHandler)
	})

	// 14. 注册 HTTP 网关处理器
	if err := srv.RegisterHTTPGateway(commonv1.RegisterHealthServiceHandlerFromEndpoint); err != nil {
		log.Fatalf("failed to register health service gateway: %v", err)
	}

	if err := srv.RegisterHTTPGateway(inventoryv1.RegisterInventoryServiceHandlerFromEndpoint); err != nil {
		log.Fatalf("failed to register cart service gateway: %v", err)
	}

	// 15. 注册 Swagger 文档
	srv.RegisterSwagger(
		server.SwaggerDoc{
			Name:        "inventory",
			FilePath:    "docs/openapi/inventory.swagger.json",
			Title:       "库存服务 API",
			Description: "库存服务 API 文档，包括库存查询、扣减、回滚等功能",
			Version:     "1.0.0",
		},
	)

	// 16. 注册中间件
	srv.UseMiddleware(
		middleware.CORS(middleware.DefaultCORSConfig()), // 1. 最外层：处理跨域
		middleware.Recovery(),                           // 2. 捕获 panic
		middleware.Logging(),                            // 3. 记录日志
		middleware.TraceID(),                            // 4. 生成 TraceID
		middleware.Auth(),                               // 5. 认证（购物车需要登录）
		middleware.CasbinRBAC(),                         // 6. RBAC 权限控制
	)

	// 17. 启动服务器（阻塞）
	if err := srv.Start(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
