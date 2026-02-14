package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"path/filepath"
	"time"
	commonv1 "zjMall/gen/go/api/proto/common"
	orderv1 "zjMall/gen/go/api/proto/order"
	"zjMall/internal/common/client"
	"zjMall/internal/common/middleware"
	"zjMall/internal/common/mq"
	registry "zjMall/internal/common/register"
	"zjMall/internal/common/server"
	"zjMall/internal/config"
	"zjMall/internal/database"
	"zjMall/internal/order-service/handler"
	"zjMall/internal/order-service/repository"
	"zjMall/internal/order-service/service"
	"zjMall/pkg"

	amqp "github.com/rabbitmq/amqp091-go"
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
	outboxRepo := repository.NewOrderOutboxRepository(db)

	// 3.1 初始化 RabbitMQ 延迟消息（用于订单超时）
	var delayedProducer mq.MessageProducer
	var delayedCh *amqp.Channel
	rabbitCfg := cfg.GetRabbitMQConfig()
	if rabbitCfg != nil && rabbitCfg.Host != "" {
		// 初始化延迟消息 exchange 和队列
		var err error
		delayedCh, err = database.InitRabbitMQ(rabbitCfg)
		if err != nil {
			log.Printf("⚠️ order-service RabbitMQ 延迟消息初始化失败: %v", err)
		} else {
			// 初始化延迟消息 exchange
			delayedExchange := "order.timeout.delayed"
			delayedQueue := "order.timeout.queue"
			if err := database.InitDelayedExchange(delayedCh, delayedExchange, delayedQueue); err != nil {
				log.Printf("⚠️ order-service 延迟消息 Exchange 初始化失败: %v", err)
			} else {
				// 开启 Publisher Confirm，确保延迟消息被 broker 正确接收
				confirmCh, err := database.EnablePublisherConfirm(delayedCh)
				if err != nil {
					log.Printf("⚠️ order-service Publisher Confirm 开启失败: %v，将使用普通生产者", err)
					delayedProducer = mq.NewMessageProducer(delayedCh, delayedQueue)
				} else {
					delayedProducer = mq.NewMessageProducerWithConfirm(delayedCh, delayedQueue, confirmCh)
				}
				log.Printf("✅ order-service 延迟消息 Exchange 初始化成功: Exchange=%s, Queue=%s", delayedExchange, delayedQueue)
			}
		}
	}

	orderService := service.NewOrderService(orderRepo, outboxRepo, productClient, inventoryClient, userClient, cartClient, redisClient, delayedProducer)
	orderHandler := handler.NewOrderServiceHandler(orderService)

	// 启动订单超时消息消费者（在 orderService 创建之后）
	if delayedCh != nil && delayedProducer != nil {
		timeoutConsumerCtx, timeoutConsumerCancel := context.WithCancel(context.Background())
		defer timeoutConsumerCancel()
		go service.StartOrderTimeoutConsumer(timeoutConsumerCtx, orderService, delayedCh, "order.timeout.queue")
		log.Println("✅ 订单超时消息消费者已启动")
	} else {
		log.Println("⚠️ 订单超时消息消费者未启动（延迟消息未初始化），将依赖补偿机制定期扫描超时订单")
	}

	// 启动订单超时补偿机制（定期扫描超时订单，作为延迟消息的兜底方案）
	// 注意：无论延迟消息是否启动，补偿机制都应该运行，确保即使延迟消息失败也能处理超时订单
	compensationCtx, compensationCancel := context.WithCancel(context.Background())
	defer compensationCancel()
	go service.StartOrderTimeoutCompensation(compensationCtx, orderService, 30*time.Minute) // 每30分钟扫描一次
	log.Println("✅ 订单超时补偿机制已启动（每30分钟扫描一次）")

	// 3.2 初始化 RabbitMQ 并启动支付成功事件消费者（可选）
	if rabbitCfg != nil && rabbitCfg.Host != "" {
		// 复制一份配置，使用单独的队列用于支付成功事件
		localCfg := *rabbitCfg
		localCfg.Queue = "payment.success.notify"

		ch, err := database.InitRabbitMQ(&localCfg)
		if err != nil {
			log.Printf("⚠️ order-service RabbitMQ 初始化失败，将跳过支付事件消费: %v", err)
		} else {
			defer database.CloseRabbitMQ()
			consumerCtx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go service.StartPaymentEventConsumer(consumerCtx, orderService, ch, localCfg.Queue)

			// 初始化outbox消息生产者（用于发送outbox事件）
			outboxCfg := *rabbitCfg
			outboxCfg.Queue = "order.outbox" // outbox事件队列
			outboxCh, err := database.InitRabbitMQ(&outboxCfg)
			if err != nil {
				log.Printf("⚠️ order-service RabbitMQ Outbox 初始化失败: %v", err)
			} else {
				outboxProducer := mq.NewMessageProducer(outboxCh, outboxCfg.Queue)
				log.Printf("✅ order-service RabbitMQ Outbox 初始化成功，队列=%s", outboxCfg.Queue)

				// 启动outbox dispatcher（定期发送outbox事件）
				dispatchCtx, dispatchCancel := context.WithCancel(context.Background())
				defer dispatchCancel()
				go func() {
					ticker := time.NewTicker(100 * time.Second) // 每5秒检查一次
					defer ticker.Stop()
					for {
						select {
						case <-dispatchCtx.Done():
							log.Println("ℹ️ Outbox 派发协程退出")
							return
						case <-ticker.C:
							if err := orderService.DispatchOutboxEvents(dispatchCtx, outboxProducer, 100); err != nil {
								log.Printf("⚠️ Outbox 派发失败: %v", err)
							}
						}
					}
				}()
			}
		}
	} else {
		log.Println("ℹ️ RabbitMQ 未配置或主机为空，订单服务将不消费支付成功事件")
	}

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
