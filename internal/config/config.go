package config

import (
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"gopkg.in/yaml.v3"
)

type ServiceConfig struct {
	GRPC struct {
		Port int `yaml:"port"`
	} `yaml:"grpc"`
	HTTP struct {
		Port int `yaml:"port"`
	} `yaml:"http"`
}

type DatabaseConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	Username        string        `yaml:"username"`
	Password        string        `yaml:"password"`
	DBName          string        `yaml:"db_name"`
	Charset         string        `yaml:"charset"`
	ParseTime       bool          `yaml:"parseTime"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
}

type RedisConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Password     string `yaml:"password"`
	DB           int    `yaml:"db"`
	PoolSize     int    `yaml:"poolSize"`
	MinIdleConns int    `yaml:"minIdleConns"`
	DialTimeout  int    `yaml:"dialTimeout"`
	ReadTimeout  int    `yaml:"readTimeout"`
	WriteTimeout int    `yaml:"writeTimeout"`
}

type RabbitMQConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	VHost    string `yaml:"vhost"`
	Queue    string `yaml:"queue"`
}

type SMSConfig struct {
	CodeLength   int `yaml:"code_length"`
	ExpireTime   int `yaml:"expire_time"`
	SendInterval int `yaml:"send_interval"`
	MaxSendCount int `yaml:"max_send_count"`
}

type JWTConfig struct {
	Secret              string        `yaml:"secret"`
	ExpiresIn           time.Duration `yaml:"expires_in"`
	RememberMeExpiresIn time.Duration `yaml:"remember_me_expires_in"`
}

type OSSConfig struct {
	Endpoint        string `yaml:"endpoint"`          // OSS 访问端点
	AccessKeyID     string `yaml:"access_key_id"`     // AccessKey ID
	AccessKeySecret string `yaml:"access_key_secret"` // AccessKey Secret
	BucketName      string `yaml:"bucket_name"`       // Bucket 名称
	BaseURL         string `yaml:"base_url"`          // 访问的基础 URL（CDN 地址或 OSS 地址）
	AvatarPath      string `yaml:"avatar_path"`       // 头像存储路径前缀
}

type ElasticsearchConfig struct {
	Host string `yaml:"host"`
}

type ServiceClientConfig struct {
	ProductServiceAddr   string `yaml:"product_service_addr"`   // 商品服务 gRPC 地址，例如 "localhost:50053"
	OrderServiceAddr     string `yaml:"order_service_addr"`     // 订单服务 gRPC 地址，例如 "localhost:50054"
	InventoryServiceAddr string `yaml:"inventory_service_addr"` // 库存服务 gRPC 地址，例如 "localhost:50055"
	UserServiceAddr      string `yaml:"user_service_addr"`      // 用户服务 gRPC 地址，例如 "localhost:50052"
	CartServiceAddr      string `yaml:"cart_service_addr"`      // 购物车服务 gRPC 地址，例如 "localhost:50054"
}

type NacosConfig struct {
	Host      string `yaml:"host"`
	Port      uint64 `yaml:"port"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	Namespace string `yaml:"namespace"`
}
type Config struct {
	Services         map[string]ServiceConfig `yaml:"services"`
	MySQL            DatabaseConfig           `yaml:"mysql"`
	ServiceDatabases map[string]string        `yaml:"service_databases"` // 服务名到数据库名的映射
	Redis            RedisConfig              `yaml:"redis"`
	SMS              SMSConfig                `yaml:"sms"`
	JWT              JWTConfig                `yaml:"jwt"`
	OSS              OSSConfig                `yaml:"oss"`
	Elasticsearch    ElasticsearchConfig      `yaml:"elasticsearch"`
	ServiceClients   ServiceClientConfig      `yaml:"service_clients"` // 服务客户端配置
	Nacos            NacosConfig              `yaml:"nacos"`
	RabbitMQ         RabbitMQConfig           `yaml:"rabbitmq"`
}

// globalConfig 持有当前生效的配置，用于 ListenConfig 动态更新。
var globalConfig atomic.Value

func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Println("Error reading config file:", err)
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Println("Error parsing config file:", err)
		return nil, err
	}
	return &config, nil
}

// LoadConfigFromNacos 先使用本地 configPath 里的 nacos 段获取 Nacos 连接信息，
// 然后从 Nacos 配置中心拉取完整业务配置。
// dataID / group 需要与你在 Nacos 控制台中创建的配置保持一致。
func LoadConfigFromNacos(configPath, dataID, group string) (*Config, error) {
	// 1. 先读本地配置文件，只为拿到 Nacos 连接信息
	localCfg, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("load local config for nacos failed: %w", err)
	}

	nacosCfg := localCfg.GetNacosConfig()

	// 2. 构造 Nacos ConfigClient
	sc := []constant.ServerConfig{
		*constant.NewServerConfig(nacosCfg.Host, nacosCfg.Port),
	}
	cc := *constant.NewClientConfig(
		constant.WithNamespaceId(nacosCfg.Namespace),
		constant.WithUsername(nacosCfg.Username),
		constant.WithPassword(nacosCfg.Password),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
	)

	configClient, err := clients.NewConfigClient(vo.NacosClientParam{
		ClientConfig:  &cc,
		ServerConfigs: sc,
	})
	if err != nil {
		return nil, fmt.Errorf("create nacos config client failed: %w", err)
	}

	if group == "" {
		group = "DEFAULT_GROUP"
	}

	// 3. 从 Nacos 拉取远程配置内容
	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: dataID,
		Group:  group,
	})
	if err != nil {
		return nil, fmt.Errorf("get config from nacos failed: %w", err)
	}
	if content == "" {
		return nil, fmt.Errorf("empty config content from nacos, dataID=%s, group=%s", dataID, group)
	}

	// 4. 解析为 Config 结构体
	var remoteCfg Config
	if err := yaml.Unmarshal([]byte(content), &remoteCfg); err != nil {
		return nil, fmt.Errorf("unmarshal nacos config failed: %w", err)
	}

	// 补回本地的 Nacos 段，方便后续代码继续使用 GetNacosConfig()
	remoteCfg.Nacos = *nacosCfg

	// 设置全局配置（供后续动态获取）
	globalConfig.Store(&remoteCfg)

	// 5. 启动 ListenConfig，监听配置变更并实时更新 globalConfig
	err = configClient.ListenConfig(vo.ConfigParam{
		DataId: dataID,
		Group:  group,
		OnChange: func(namespace, group, dataId, data string) {
			log.Printf("[Nacos] config changed, dataId=%s, group=%s", dataId, group)
			var updatedCfg Config
			if err := yaml.Unmarshal([]byte(data), &updatedCfg); err != nil {
				log.Printf("[Nacos] unmarshal updated config failed: %v", err)
				return
			}
			updatedCfg.Nacos = *nacosCfg
			globalConfig.Store(&updatedCfg)
		},
	})
	if err != nil {
		log.Printf("[Nacos] ListenConfig failed (will still use current config): %v", err)
	}

	return &remoteCfg, nil
}

// GetDynamicConfig 返回最近一次从 Nacos 拉取/监听到的配置。
// 对于需要“实时”读取配置的业务逻辑，可以优先使用该方法。
func GetDynamicConfig() *Config {
	if v := globalConfig.Load(); v != nil {
		if cfg, ok := v.(*Config); ok {
			return cfg
		}
	}
	return nil
}

func (c *Config) GetServiceConfig(serviceName string) (*ServiceConfig, error) {
	serviceConfig, exists := c.Services[serviceName]
	if !exists {
		log.Println("Service not found in config:", serviceName)
		return nil, fmt.Errorf("service %s not found in config", serviceName)
	}
	return &serviceConfig, nil
}

func (c *Config) GetDatabaseConfigForService(serviceName string) (*DatabaseConfig, error) {
	dbName, exists := c.ServiceDatabases[serviceName]
	if !exists {
		// 如果没有配置，使用命名约定：service_name -> service_name_db
		dbName = serviceName + "_db"
	}

	config := c.MySQL
	config.DBName = dbName
	return &config, nil
}

func (c *Config) GetRedisConfig() *RedisConfig {
	return &c.Redis
}

func (c *Config) GetSMSConfig() *SMSConfig {
	return &c.SMS
}

func (c *Config) GetJWTConfig() *JWTConfig {
	return &c.JWT
}

func (c *Config) GetOSSConfig() *OSSConfig {
	return &c.OSS
}

func (c *Config) GetElasticsearchConfig() *ElasticsearchConfig {
	return &c.Elasticsearch
}

func (c *Config) GetServiceClientsConfig() *ServiceClientConfig {
	return &c.ServiceClients
}

func (c *Config) GetNacosConfig() *NacosConfig {
	return &c.Nacos
}

func (c *Config) GetRabbitMQConfig() *RabbitMQConfig {
	return &c.RabbitMQ
}
