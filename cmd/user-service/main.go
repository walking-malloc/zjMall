package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
	userv1 "zjMall/gen/go/api/proto/user"
	"zjMall/internal/common/cache"
	"zjMall/internal/common/middleware"
	"zjMall/internal/config"
	"zjMall/internal/database"
	"zjMall/internal/sms"
	"zjMall/internal/user-service/handler"
	"zjMall/internal/user-service/repository"
	"zjMall/internal/user-service/service"
	"zjMall/pkg/validator"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	// 创建gRPC服务器
	grpcServer := grpc.NewServer()

	// 注册user服务
	userv1.RegisterUserServiceServer(grpcServer, userServiceHandler)

	// 启动 gRPC 服务器（在 goroutine 中，避免阻塞）
	grpcAddr := fmt.Sprintf(":%d", serviceCfg.GRPC.Port)

	//创建http服务器
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 创建 HTTP 路由
	mux := runtime.NewServeMux()

	// 连接到 gRPC 服务器
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// 注册用户服务的 HTTP 网关处理器
	err = userv1.RegisterUserServiceHandlerFromEndpoint(
		ctx, mux, grpcAddr, opts,
	)
	if err != nil {
		log.Fatalf("failed to register user service gateway: %v", err)
	}

	handler := middleware.Chain(
		middleware.Recovery(),
		middleware.CORS(middleware.DefaultCORSConfig()),
		middleware.TraceID(),
		middleware.Logging(),
	)(mux)

	// 启动 HTTP 服务器
	httpAddr := fmt.Sprintf(":%d", serviceCfg.HTTP.Port)
	httpServer := &http.Server{
		Addr:         httpAddr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	startGRPCServer(grpcServer, grpcAddr)
	startHTTPServer(httpServer, httpAddr)
	gracefulShutdown(httpServer, grpcServer)

}

func startGRPCServer(grpcServer *grpc.Server, grpcAddr string) {
	grpcLis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", grpcAddr, err)
	}

	go func() {
		log.Printf("gRPC server listening on %s", grpcAddr)
		if err := grpcServer.Serve(grpcLis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()
}

func startHTTPServer(httpServer *http.Server, httpAddr string) {
	go func() {
		log.Printf("HTTP server listening on %s", httpAddr)
		log.Printf("Try: http://localhost%s/healthz", httpAddr)

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to serve HTTP: %v", err)
		}
	}()
}

func gracefulShutdown(httpServer *http.Server, grpcServer *grpc.Server) {
	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, syscall.SIGINT, syscall.SIGTERM) //当收到SIGINT（Ctrl+C）或SIGTERM（终止信号（通常是系统关闭、K8s 停止容器时发送））信号时，关闭服务器
	<-signChan

	// ========== 新增：优雅关闭 HTTP 服务器 ==========
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	log.Println("正在关闭 HTTP 服务器...")
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP 服务器关闭错误: %v", err)
	} else {
		log.Println("HTTP 服务器已关闭")
	}

	// ========== 新增：优雅关闭 gRPC 服务器 ==========
	log.Println("正在关闭 gRPC 服务器...")
	grpcServer.GracefulStop()
	log.Println("gRPC 服务器已关闭")

	log.Println("所有服务器已关闭")
}
