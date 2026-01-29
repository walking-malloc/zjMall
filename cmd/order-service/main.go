package main

import (
	"fmt"
	"log"
	"math/rand"
	"path/filepath"
	"time"
	commonv1 "zjMall/gen/go/api/proto/common"
	orderv1 "zjMall/gen/go/api/proto/order"
	"zjMall/internal/common/middleware"
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

	// 初始化 JWT（如果订单需要鉴权）
	pkg.InitJWT(cfg.GetJWTConfig())

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
	orderService := service.NewOrderService(orderRepo)
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
