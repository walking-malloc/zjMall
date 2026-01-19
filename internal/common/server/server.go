package server

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"zjMall/internal/common/middleware"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Config 服务器配置
type Config struct {
	GRPCAddr string
	HTTPAddr string
}

// Server 通用服务器
type Server struct {
	grpcServer *grpc.Server
	httpServer *http.Server
	gwMux      *runtime.ServeMux
	httpMux    *http.ServeMux
	config     *Config
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewServer 创建新的服务器实例
func NewServer(cfg *Config) *Server {
	ctx, cancel := context.WithCancel(context.Background())

	// 配置 gRPC Gateway 的 metadata 传递
	// 将 HTTP context 中的 user_id 传递到 gRPC metadata
	// 注意：这里使用字符串 "user_id" 作为 key，与 middleware.UserIDKey 的值一致
	gwMux := runtime.NewServeMux(
		runtime.WithMetadata(func(ctx context.Context, req *http.Request) metadata.MD {
			md := metadata.MD{}
			// 从 HTTP context 中获取 user_id（由认证中间件设置），并传递到 gRPC metadata
			// middleware.UserIDKey 的值是 "user_id"
			if userID := ctx.Value("user_id"); userID != nil {
				if userIDStr, ok := userID.(string); ok && userIDStr != "" {
					md.Set("user_id", userIDStr)
					log.Printf("gRPC Gateway: 传递 user_id 到 metadata: %s", userIDStr)
				}
			}
			return md
		}),
	)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(middleware.UnaryAuthInterceptor))
	s := &Server{
		grpcServer: grpcServer,
		gwMux:      gwMux,
		httpMux:    http.NewServeMux(),
		config:     cfg,
		ctx:        ctx,
		cancel:     cancel,
	}

	// 将 gRPC-Gateway mux 注册到主 HTTP mux
	s.httpMux.Handle("/", s.gwMux)

	return s
}

// RegisterGRPCService 注册 gRPC 服务
// registerFunc 是一个函数，接收 *grpc.Server 并注册服务
func (s *Server) RegisterGRPCService(registerFunc func(*grpc.Server)) {
	registerFunc(s.grpcServer)
}

// RegisterHTTPGateway 注册 HTTP 网关处理器
// registerFunc 是 gRPC-Gateway 生成的注册函数
func (s *Server) RegisterHTTPGateway(registerFunc func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error) error {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	return registerFunc(s.ctx, s.gwMux, s.config.GRPCAddr, opts)
}

// AddRoute 添加自定义 HTTP 路由（用于 Swagger 等）
func (s *Server) AddRoute(pattern string, handler http.HandlerFunc) {
	s.httpMux.HandleFunc(pattern, handler)
}

// SwaggerDoc 表示一个 Swagger 文档配置
type SwaggerDoc struct {
	Name        string // 文档名称，如 "user", "health"
	FilePath    string // 文档文件路径，如 "docs/openapi/user.swagger.json"
	Title       string // 文档标题，如 "用户服务 API"
	Description string // 文档描述
	Version     string // 文档版本，如 "1.0.0"
}

// RegisterSwagger 注册 Swagger 文档路由
// docs: Swagger 文档列表，第一个文档作为默认文档
func (s *Server) RegisterSwagger(docs ...SwaggerDoc) {
	if len(docs) == 0 {
		return
	}

	// 注册每个文档的 JSON 路由
	for _, doc := range docs {
		doc := doc // 避免闭包问题
		s.httpMux.HandleFunc("/swagger/"+doc.Name+".json", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			// 如果设置了自定义元数据，修改 JSON 后返回
			if doc.Title != "" || doc.Description != "" || doc.Version != "" {
				data, err := ioutil.ReadFile(doc.FilePath)
				if err != nil {
					http.Error(w, "Failed to read swagger file", http.StatusInternalServerError)
					return
				}

				var swagger map[string]interface{}
				if err := json.Unmarshal(data, &swagger); err != nil {
					http.Error(w, "Failed to parse swagger file", http.StatusInternalServerError)
					return
				}

				// 更新 info 信息
				info, ok := swagger["info"].(map[string]interface{})
				if !ok {
					info = make(map[string]interface{})
					swagger["info"] = info
				}

				if doc.Title != "" {
					info["title"] = doc.Title
				}
				if doc.Description != "" {
					info["description"] = doc.Description
				}
				if doc.Version != "" {
					info["version"] = doc.Version
				}

				// 返回修改后的 JSON
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(swagger)
				return
			}

			// 如果没有自定义元数据，直接返回原文件
			http.ServeFile(w, r, doc.FilePath)
		})
	}

	// 默认 Swagger UI 重定向到第一个文档
	defaultDoc := docs[0]
	s.httpMux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		swaggerURL := scheme + "://" + r.Host + "/swagger/" + defaultDoc.Name + ".json"
		redirectURL := "https://petstore.swagger.io/?url=" + swaggerURL
		http.Redirect(w, r, redirectURL, http.StatusFound)
	})

	// 兼容旧的路由，重定向到默认文档
	s.httpMux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/"+defaultDoc.Name+".json", http.StatusFound)
	})
}

// UseMiddleware 使用中间件
func (s *Server) UseMiddleware(middlewares ...func(http.Handler) http.Handler) {
	var handler http.Handler = s.httpMux
	for _, mw := range middlewares {
		handler = mw(handler)
	}
	s.httpServer = &http.Server{
		Addr:         s.config.HTTPAddr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// Start 启动服务器（阻塞）
func (s *Server) Start() error {
	// 启动 gRPC 服务器
	go s.startGRPC()

	// 启动 HTTP 服务器
	go s.startHTTP()

	// 等待中断信号
	s.gracefulShutdown()

	return nil
}

// startGRPC 启动 gRPC 服务器
func (s *Server) startGRPC() {
	lis, err := net.Listen("tcp", s.config.GRPCAddr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", s.config.GRPCAddr, err)
	}

	log.Printf("gRPC server listening on %s", s.config.GRPCAddr)
	if err := s.grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}

// startHTTP 启动 HTTP 服务器
func (s *Server) startHTTP() {
	log.Printf("HTTP server listening on %s", s.config.HTTPAddr)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to serve HTTP: %v", err)
	}
}

// gracefulShutdown 优雅关闭
func (s *Server) gracefulShutdown() {
	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, syscall.SIGINT, syscall.SIGTERM)
	<-signChan

	log.Println("正在关闭服务器...")

	// 关闭上下文
	s.cancel()

	// 优雅关闭 HTTP 服务器
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	log.Println("正在关闭 HTTP 服务器...")
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP 服务器关闭错误: %v", err)
	} else {
		log.Println("HTTP 服务器已关闭")
	}

	// 优雅关闭 gRPC 服务器
	log.Println("正在关闭 gRPC 服务器...")
	s.grpcServer.GracefulStop()
	log.Println("gRPC 服务器已关闭")

	log.Println("所有服务器已关闭")
}
