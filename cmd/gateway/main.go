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
	commonv1 "zjMall/gen/go/api/proto/common"
	"zjMall/internal/common/handler"
	"zjMall/internal/config"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	configPath := filepath.Join("./configs", "config.yaml")
	config, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	serviceName := "gateway"
	serviceCfg, err := config.GetServiceConfig(serviceName)
	if err != nil {
		log.Fatalf("Error getting service config: %v", err)
	}

	// 创建gRPC服务器
	grpcServer := grpc.NewServer()

	// 注册健康检查服务
	healthHandler := handler.NewHealthHandler()
	commonv1.RegisterHealthServiceServer(grpcServer, healthHandler)

	// 启动 gRPC 服务器（在 goroutine 中，避免阻塞）
	grpcAddr := fmt.Sprintf(":%d", serviceCfg.GRPC.Port)
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

	// 注册 HTTP 网关处理器
	err = commonv1.RegisterHealthServiceHandlerFromEndpoint(
		ctx, mux, grpcAddr, opts,
	)
	if err != nil {
		log.Fatalf("failed to register gateway: %v", err)
	}

	// 启动 HTTP 服务器
	httpAddr := fmt.Sprintf(":%d", serviceCfg.HTTP.Port)
	httpServer := &http.Server{
		Addr:         httpAddr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	go func() {
		log.Printf("HTTP server listening on %s", httpAddr)
		log.Printf("Try: http://localhost%s/healthz", httpAddr)

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to serve HTTP: %v", err)
		}
	}()

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

	log.Println("所有服务器已优雅关闭")

}
