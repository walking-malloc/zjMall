package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	commonv1 "zjMall/gen/go/api/proto/common"
	productv1 "zjMall/gen/go/api/proto/product"
	"zjMall/internal/common/authz"
	"zjMall/internal/common/cache"
	"zjMall/internal/common/middleware"
	registry "zjMall/internal/common/register"
	"zjMall/internal/common/server"
	"zjMall/internal/config"
	"zjMall/internal/database"
	"zjMall/internal/product-service/handler"
	"zjMall/internal/product-service/repository"
	"zjMall/internal/product-service/service"
	"zjMall/pkg"
	"zjMall/pkg/validator"

	"google.golang.org/grpc"
)

const serviceName = "product-service"
const serviceIP = "127.0.0.1"

// todo 需要改为商品服务的配置
func main() {
	logFile, err := pkg.InitLog(serviceName)
	if err != nil {
		log.Fatalf("Error initializing log: %v", err)
	}
	defer logFile.Close()
	log.Printf("==== %s starting ====", serviceName)

	//1.加载配置
	log.Println("📝 加载配置文件...")
	configPath := filepath.Join("./configs", "config.yaml")
	// 从 Nacos 配置中心加载业务配置（DataID/Group 需要与你在 Nacos 中保持一致）
	config, err := config.LoadConfigFromNacos(configPath, "zjmall-dev.yaml", "DEFAULT_GROUP")
	if err != nil {
		log.Fatalf("❌ 从 Nacos 加载配置失败: %v", err)
	}
	log.Println("✅ 配置文件加载成功")
	// 加载完配置 cfg 之后：
	if err := authz.InitCasbin(); err != nil {
		log.Fatalf("❌ Casbin 初始化失败: %v", err)
	}
	//2.初始化Nacos
	svcCfg, _ := config.GetServiceConfig(serviceName)
	nacosConfig := config.GetNacosConfig()
	nacosClient, err := registry.NewNacosNamingClient(nacosConfig)
	if err != nil {
		log.Fatalf("❌ Nacos 初始化失败: %v", err)
	}
	registry.RegisterService(nacosClient, serviceName, serviceIP, uint64(svcCfg.GRPC.Port))
	//初始化JWT
	log.Println("🔧 初始化 JWT...")
	jwtConfig := config.GetJWTConfig()
	pkg.InitJWT(jwtConfig)
	log.Println("✅ JWT 初始化成功")

	//2.初始化数据库（使用服务特定的数据库配置）
	serviceName := "product-service"
	log.Printf("🔧 初始化数据库连接 (服务: %s)...", serviceName)
	mysqlConfig, err := config.GetDatabaseConfigForService(serviceName)
	if err != nil {
		log.Fatalf("❌ 获取数据库配置失败 (%s): %v", serviceName, err)
	}
	db, err := database.InitMySQL(mysqlConfig)
	if err != nil {
		log.Fatalf("❌ MySQL 初始化失败: %v", err)
	}
	defer database.CloseMySQL()
	log.Println("✅ MySQL 连接成功")

	//3.初始化redis
	log.Println("🔧 初始化 Redis 连接...")
	redisConfig := config.GetRedisConfig()
	redisClient, err := database.InitRedis(redisConfig)
	if err != nil {
		log.Fatalf("❌ Redis 初始化失败: %v", err)
	}
	defer database.CloseRedis()
	log.Println("✅ Redis 连接成功")

	//4.初始化elasticsearch
	log.Println("🔧 初始化 Elasticsearch 连接...")
	elasticsearchConfig := config.GetElasticsearchConfig()
	elasticsearchClient, err := database.NewElasticsearchClient(elasticsearchConfig)
	if err != nil {
		log.Fatalf("❌ Elasticsearch 初始化失败: %v", err)
	}
	defer elasticsearchClient.Close(context.Background())
	log.Println("✅ Elasticsearch 连接成功")

	//4.初始化校验器
	log.Println("🔧 初始化校验器...")
	validator.Init()
	log.Println("✅ 校验器初始化成功")

	// 6. 创建仓库
	log.Println("🔧 创建 Repository...")
	cacheRepo := cache.NewCacheRepository(redisClient)
	categoryRepo := repository.NewCategoryRepository(db, cacheRepo)
	brandRepo := repository.NewBrandRepository(db, cacheRepo)
	productRepo := repository.NewProductRepository(db, cacheRepo)
	tagRepo := repository.NewTagRepository(db, cacheRepo)
	skuRepo := repository.NewSkuRepository(db)
	attributeRepo := repository.NewAttributeRepository(db)
	attributeValueRepo := repository.NewAttributeValueRepository(db)

	// 创建 ES 搜索仓库
	searchRepo := repository.NewSearchRepository(elasticsearchClient.GetClient())
	log.Println("✅ Repository 创建成功")

	// 7. 初始化 ES 索引
	log.Println("🔧 初始化 Elasticsearch 索引...")
	if err := searchRepo.CreateIndex(context.Background()); err != nil {
		log.Printf("⚠️  创建 ES 索引失败（可能已存在）: %v", err)
	} else {
		log.Println("✅ Elasticsearch 索引创建成功")
	}

	// 8. 创建搜索服务
	log.Println("🔧 创建 SearchService...")
	searchService := service.NewSearchService(
		searchRepo,
		productRepo,
		categoryRepo,
		brandRepo,
		tagRepo,
		attributeRepo,
		attributeValueRepo,
		skuRepo,
	)
	log.Println("✅ SearchService 创建成功")

	// 9. 创建Service
	log.Println("🔧 创建 Service...")
	productService := service.NewProductService(categoryRepo, brandRepo, productRepo, tagRepo, skuRepo, attributeRepo, attributeValueRepo, searchService)
	log.Println("✅ Service 创建成功")

	//7.创建Handler
	log.Println("🔧 创建 Handler...")
	productServiceHandler := handler.NewProductServiceHandler(productService)
	log.Println("✅ Handler 创建成功")

	log.Println("🔧 获取服务配置...")
	serviceCfg, err := config.GetServiceConfig(serviceName)
	if err != nil {
		log.Fatalf("❌ 获取服务配置失败: %v", err)
	}
	log.Printf("✅ 服务配置获取成功 (gRPC: :%d, HTTP: :%d)", serviceCfg.GRPC.Port, serviceCfg.HTTP.Port)

	// 创建服务器实例
	srv := server.NewServer(&server.Config{
		GRPCAddr: fmt.Sprintf(":%d", serviceCfg.GRPC.Port),
		HTTPAddr: fmt.Sprintf(":%d", serviceCfg.HTTP.Port),
	})

	// 注册 gRPC 服务
	srv.RegisterGRPCService(func(grpcServer *grpc.Server) {
		// 注册用户服务
		productv1.RegisterProductServiceServer(grpcServer, productServiceHandler)
	})

	// 注册自定义HTTP路由（头像上传）- 必须在 gRPC-Gateway 之前注册，确保优先匹配
	// srv.AddRoute("/api/v1/users/avatar", productServiceHandler.UploadAvatarHTTP)

	// 注册 HTTP 网关处理器
	if err := srv.RegisterHTTPGateway(commonv1.RegisterHealthServiceHandlerFromEndpoint); err != nil {
		log.Fatalf("failed to register health service gateway: %v", err)
	}

	if err := srv.RegisterHTTPGateway(productv1.RegisterProductServiceHandlerFromEndpoint); err != nil {
		log.Fatalf("failed to register user service gateway: %v", err)
	}
	// 注册 Swagger 文档
	srv.RegisterSwagger(
		server.SwaggerDoc{
			Name:        "user",
			FilePath:    "docs/openapi/product.swagger.json",
			Title:       "商品服务 API",
			Description: "商品服务 API 文档，包括商品类目、品牌、商品、SKU等功能",
			Version:     "1.0.0",
		},
	)

	srv.UseMiddleware(
		middleware.CORS(middleware.DefaultCORSConfig()), // 1. 最外层：处理跨域（所有响应都需要）
		middleware.Recovery(),                           // 2. 捕获 panic（需要 TraceID）
		middleware.Logging(),                            // 3. 记录日志（需要 TraceID）
		middleware.TraceID(),                            // 4. 生成 TraceID（供 Logging 和 Recovery 使用）
		middleware.Auth(),                               // 5. 认证中间件：验证 token 并注入 user_id 到 context
		middleware.CasbinRBAC(),                         // 6. Casbin RBAC 鉴权
	)

	// 启动服务器（阻塞）
	log.Println("🚀 启动服务器...")
	if err := srv.Start(); err != nil {
		log.Fatalf("❌ 服务器启动失败: %v", err)
	}
}
