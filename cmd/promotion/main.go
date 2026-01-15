package main

import (
	"fmt"
	"log"
	"path/filepath"
	commonv1 "zjMall/gen/go/api/proto/common"
	promotionv1 "zjMall/gen/go/api/proto/promotion"
	"zjMall/internal/common/cache"
	"zjMall/internal/common/middleware"
	"zjMall/internal/common/server"
	"zjMall/internal/config"
	"zjMall/internal/database"
	"zjMall/internal/promotion-service/handler"
	"zjMall/internal/promotion-service/repository"
	"zjMall/internal/promotion-service/service"
	"zjMall/pkg/validator"

	"google.golang.org/grpc"
)

func main() {
	log.Println("🚀 开始启动促销服务...")

	// 1. 加载配置
	configPath := filepath.Join("./configs", "config.yaml")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("❌ 加载配置失败: %v", err)
	}

	// 2. 初始化数据库
	serviceName := "promotion-service"
	mysqlConfig, err := cfg.GetDatabaseConfigForService(serviceName)
	if err != nil {
		log.Fatalf("❌ 获取数据库配置失败: %v", err)
	}
	db, err := database.InitMySQL(mysqlConfig)
	if err != nil {
		log.Fatalf("❌ MySQL 初始化失败: %v", err)
	}
	defer database.CloseMySQL()

	// 3. 初始化 Redis
	redisConfig := cfg.GetRedisConfig()
	redisClient, err := database.InitRedis(redisConfig)
	if err != nil {
		log.Fatalf("❌ Redis 初始化失败: %v", err)
	}
	defer database.CloseRedis()

	// 4. 初始化校验器
	validator.Init()

	// 5. 创建 Repository
	cacheRepo := cache.NewCacheRepository(redisClient)
	promotionRepo := repository.NewPromotionRepository(db, cacheRepo)
	couponRepo := repository.NewCouponRepository(db, cacheRepo)

	// 6. 创建 Service
	promotionService := service.NewPromotionService(promotionRepo, couponRepo)

	// 7. 创建 Handler
	promotionHandler := handler.NewPromotionServiceHandler(promotionService)

	// 8. 获取服务配置
	serviceCfg, err := cfg.GetServiceConfig(serviceName)
	if err != nil {
		log.Fatalf("❌ 获取服务配置失败: %v", err)
	}

	// 9. 创建服务器
	srv := server.NewServer(&server.Config{
		GRPCAddr: fmt.Sprintf(":%d", serviceCfg.GRPC.Port),
		HTTPAddr: fmt.Sprintf(":%d", serviceCfg.HTTP.Port),
	})

	// 10. 注册 gRPC 服务
	srv.RegisterGRPCService(func(grpcServer *grpc.Server) {
		commonv1.RegisterHealthServiceServer(grpcServer, nil) // 健康检查
		promotionv1.RegisterPromotionServiceServer(grpcServer, promotionHandler)
	})

	// 11. 注册 HTTP 网关
	if err := srv.RegisterHTTPGateway(commonv1.RegisterHealthServiceHandlerFromEndpoint); err != nil {
		log.Fatalf("❌ 注册健康检查网关失败: %v", err)
	}
	if err := srv.RegisterHTTPGateway(promotionv1.RegisterPromotionServiceHandlerFromEndpoint); err != nil {
		log.Fatalf("❌ 注册促销服务网关失败: %v", err)
	}

	// 12. 注册中间件
	srv.UseMiddleware(
		middleware.CORS(middleware.DefaultCORSConfig()),
		middleware.Recovery(),
		middleware.Logging(),
		middleware.TraceID(),
	)

	// 13. 启动服务器
	log.Println("✅ 促销服务启动成功")
	if err := srv.Start(); err != nil {
		log.Fatalf("❌ 启动服务器失败: %v", err)
	}
}
