package main

import (
	"fmt"
	"log"
	"path/filepath"
	commonv1 "zjMall/gen/go/api/proto/common"
	userv1 "zjMall/gen/go/api/proto/user"
	"zjMall/internal/common/cache"
	"zjMall/internal/common/middleware"
	"zjMall/internal/common/server"
	"zjMall/internal/config"
	"zjMall/internal/database"
	"zjMall/internal/sms"
	"zjMall/internal/user-service/handler"
	"zjMall/internal/user-service/repository"
	"zjMall/internal/user-service/service"
	"zjMall/pkg/validator"

	"google.golang.org/grpc"
)

func main() {
	//1.加载配置
	configPath := filepath.Join("./configs", "config.yaml")
	config, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	//2.初始化数据库
	mysqlConfig := config.GetMySQLConfig()
	db, err := database.InitMySQL(mysqlConfig)
	if err != nil {
		log.Fatalf("Error initializing MySQL: %v", err)
	}
	defer database.CloseMySQL()

	//3.初始化redis
	redisConfig := config.GetRedisConfig()
	redisClient, err := database.InitRedis(redisConfig)
	if err != nil {
		log.Fatalf("Error initializing Redis: %v", err)
	}
	defer database.CloseRedis()

	//4.初始化校验器
	validator.Init()

	// 5. 创建通用的缓存仓库（所有服务共享）
	baseCacheRepo := cache.NewCacheRepository(redisClient)

	// 6. 创建用户仓库
	userRepo := repository.NewUserRepository(db, baseCacheRepo)

	// 7. 获取短信配置并创建短信客户端（Mock）
	smsConfig := config.GetSMSConfig()
	smsClient := sms.NewMockSMSClient()
	log.Println("✅ 使用 Mock 短信服务（学习模式）")

	// 8. 创建Service
	userService := service.NewUserService(userRepo, smsClient, *smsConfig)

	//7.创建Handler
	userServiceHandler := handler.NewUserServiceHandler(userService)

	serviceName := "user-service"
	serviceCfg, err := config.GetServiceConfig(serviceName)
	if err != nil {
		log.Fatalf("Error getting service config: %v", err)
	}

	// 创建服务器实例
	srv := server.NewServer(&server.Config{
		GRPCAddr: fmt.Sprintf(":%d", serviceCfg.GRPC.Port),
		HTTPAddr: fmt.Sprintf(":%d", serviceCfg.HTTP.Port),
	})

	// 注册 gRPC 服务
	srv.RegisterGRPCService(func(grpcServer *grpc.Server) {
		// 注册用户服务
		userv1.RegisterUserServiceServer(grpcServer, userServiceHandler)
	})

	// 注册 HTTP 网关处理器
	if err := srv.RegisterHTTPGateway(commonv1.RegisterHealthServiceHandlerFromEndpoint); err != nil {
		log.Fatalf("failed to register health service gateway: %v", err)
	}

	if err := srv.RegisterHTTPGateway(userv1.RegisterUserServiceHandlerFromEndpoint); err != nil {
		log.Fatalf("failed to register user service gateway: %v", err)
	}

	// 注册 Swagger 文档
	srv.RegisterSwagger(
		server.SwaggerDoc{
			Name:        "user",
			FilePath:    "docs/openapi/user.swagger.json",
			Title:       "用户服务 API",
			Description: "用户服务 API 文档，包括用户注册、登录、短信验证码等功能",
			Version:     "1.0.0",
		},
	)

	// 使用中间件
	srv.UseMiddleware(
		middleware.Recovery(),
		middleware.CORS(middleware.DefaultCORSConfig()),
		middleware.TraceID(),
		middleware.Logging(),
	)

	// 启动服务器（阻塞）
	if err := srv.Start(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
