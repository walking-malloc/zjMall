package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	commonv1 "zjMall/gen/go/api/proto/common"
	paymentv1 "zjMall/gen/go/api/proto/payment"
	"zjMall/internal/common/authz"
	"zjMall/internal/common/cache"
	"zjMall/internal/common/client"
	"zjMall/internal/common/lock"
	"zjMall/internal/common/middleware"
	"zjMall/internal/common/mq"
	registry "zjMall/internal/common/register"
	"zjMall/internal/common/server"
	"zjMall/internal/config"
	"zjMall/internal/database"
	"zjMall/internal/payment-service/handler"
	"zjMall/internal/payment-service/repository"
	"zjMall/internal/payment-service/service"
	"zjMall/pkg"

	"google.golang.org/grpc"
)

const serviceName = "payment-service"
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
	// 加载完配置 cfg 之后：
	if err := authz.InitCasbin(); err != nil {
		log.Fatalf("❌ Casbin 初始化失败: %v", err)
	}
	// 初始化 JWT（与用户服务、网关保持同一套 secret）
	jwtCfg := cfg.GetJWTConfig()
	pkg.InitJWT(jwtCfg)

	// 2. 初始化 Nacos 注册中心
	svcCfg, _ := cfg.GetServiceConfig(serviceName)
	nacosConfig := cfg.GetNacosConfig()
	nacosClient, err := registry.NewNacosNamingClient(nacosConfig)
	if err != nil {
		log.Fatalf("❌ Nacos 初始化失败: %v", err)
	}
	registry.RegisterService(nacosClient, serviceName, serviceIP, uint64(svcCfg.GRPC.Port))

	// 3. 初始化 MySQL（支付库）
	mysqlConfig, err := cfg.GetDatabaseConfigForService(serviceName)
	if err != nil {
		log.Fatalf("Error getting database config for %s: %v", serviceName, err)
	}
	db, err := database.InitMySQL(mysqlConfig)
	if err != nil {
		log.Fatalf("Error initializing MySQL: %v", err)
	}
	defer database.CloseMySQL()

	// 4. 初始化 Redis
	redisConfig := cfg.GetRedisConfig()
	redisClient, err := database.InitRedis(redisConfig)
	if err != nil {
		log.Fatalf("Error initializing Redis: %v", err)
	}
	defer database.CloseRedis()

	// 5. 初始化 RabbitMQ（用于发送支付成功事件）

	var paymentMQProducer mq.MessageProducer
	rabbitCfg := cfg.GetRabbitMQConfig()
	if rabbitCfg != nil && rabbitCfg.Host != "" {
		// 为支付成功事件单独使用一个队列
		localCfg := *rabbitCfg
		localCfg.Queue = "payment.success.notify"

		ch, err := database.InitRabbitMQ(&localCfg)
		if err != nil {
			log.Printf("⚠️ payment-service RabbitMQ 初始化失败，支付成功事件将不会发送到 MQ: %v", err)
		} else {
			defer database.CloseRabbitMQ()
			paymentMQProducer = mq.NewMessageProducer(ch, localCfg.Queue)
			log.Printf("✅ payment-service RabbitMQ 初始化成功，队列=%s", localCfg.Queue)
		}
	} else {
		log.Println("ℹ️ RabbitMQ 未配置或主机为空，支付服务将不发送 MQ 事件")
	}

	// 6. 初始化缓存与分布式锁
	cacheRepo := cache.NewCacheRepository(redisClient)
	lockService := lock.NewRedisLockService(redisClient)

	// 7. 初始化订单服务客户端（优先通过 Nacos 发现）
	var orderClient client.OrderClient
	orderServiceAddr, err := registry.SelectOneHealthyInstance(nacosClient, "order-service")
	if err != nil || orderServiceAddr == "" {
		// 回退到配置中的地址
		log.Printf("⚠️ 从 Nacos 发现订单服务失败，将尝试使用配置中的备用地址: %v", err)
		serviceClientsConfig := cfg.GetServiceClientsConfig()
		if serviceClientsConfig != nil && serviceClientsConfig.OrderServiceAddr != "" {
			orderServiceAddr = serviceClientsConfig.OrderServiceAddr
			log.Printf("ℹ️ 使用配置中的订单服务备用地址: %s", orderServiceAddr)
		}
	}
	if orderServiceAddr != "" {
		orderClient, err = client.NewOrderClient(orderServiceAddr)
		if err != nil {
			log.Printf("⚠️ 订单服务客户端初始化失败，部分校验功能将不可用: %v", err)
		} else {
			defer orderClient.Close()
		}
	} else {
		log.Println("⚠️ 未找到订单服务地址，支付创建将无法校验订单")
	}

	// 8. 创建仓库
	paymentRepo := repository.NewPaymentRepository(db)
	paymentLogRepo := repository.NewPaymentLogRepository(db)
	paymentChannelRepo := repository.NewPaymentChannelRepository(db)
	outboxRepo := repository.NewPaymentOutboxRepository(db)

	// 9. 创建 PaymentService
	var paymentTimeout = 30 * time.Minute
	paymentService := service.NewPaymentService(
		paymentRepo,
		paymentLogRepo,
		paymentChannelRepo,
		orderClient,
		cacheRepo,
		paymentTimeout,
		lockService,
		paymentMQProducer,
		outboxRepo,
	)

	// 10. 启动 Outbox 派发协程（定期将 Outbox 事件发送到 MQ）
	if paymentMQProducer != nil {
		dispatchCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-dispatchCtx.Done():
					log.Println("ℹ️ Outbox 派发协程退出")
					return
				case <-ticker.C:
					if err := paymentService.DispatchOutboxEvents(dispatchCtx, 100); err != nil {
						log.Printf("⚠️ Outbox 派发失败: %v", err)
					}
				}
			}
		}()
	}

	// 11. 创建 Handler
	paymentHandler := handler.NewPaymentHandler(paymentService)

	// 12. 创建服务器实例
	serviceCfg, err := cfg.GetServiceConfig(serviceName)
	if err != nil {
		log.Fatalf("Error getting service config: %v", err)
	}

	srv := server.NewServer(&server.Config{
		GRPCAddr: fmt.Sprintf(":%d", serviceCfg.GRPC.Port),
		HTTPAddr: fmt.Sprintf(":%d", serviceCfg.HTTP.Port),
	})

	// 13. 注册 gRPC 服务
	srv.RegisterGRPCService(func(grpcServer *grpc.Server) {
		paymentv1.RegisterPaymentServiceServer(grpcServer, paymentHandler)
	})

	// 14. 注册 HTTP 网关处理器
	if err := srv.RegisterHTTPGateway(commonv1.RegisterHealthServiceHandlerFromEndpoint); err != nil {
		log.Fatalf("failed to register health service gateway: %v", err)
	}
	if err := srv.RegisterHTTPGateway(paymentv1.RegisterPaymentServiceHandlerFromEndpoint); err != nil {
		log.Fatalf("failed to register payment service gateway: %v", err)
	}

	// 15. 注册 Swagger 文档
	srv.RegisterSwagger(
		server.SwaggerDoc{
			Name:        "payment",
			FilePath:    "docs/openapi/payment.swagger.json",
			Title:       "支付服务 API",
			Description: "支付服务 API 文档，包括创建支付单、支付回调、支付状态查询等功能",
			Version:     "1.0.0",
		},
	)

	// 16. 注册中间件
	srv.UseMiddleware(
		middleware.CORS(middleware.DefaultCORSConfig()),
		middleware.Recovery(),
		middleware.Logging(),
		middleware.TraceID(),
		middleware.Auth(),
		middleware.CasbinRBAC(),
	)

	// 17. 启动服务器
	if err := srv.Start(); err != nil {
		log.Fatalf("failed to start payment-service: %v", err)
	}
}
