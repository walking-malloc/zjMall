package config

import (
	"fmt"
	"log"
	"os"
	"time"

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

type RocketMQConfig struct {
	// RocketMQ 5.x 使用 Proxy 的 gRPC Endpoint（不再是 NameServer）
	Endpoint      string            `yaml:"endpoint"`        // Proxy gRPC Endpoint（例如 "127.0.0.1:8081"）
	AccessKey     string            `yaml:"access_key"`      // 访问密钥（可选，如果启用了认证）
	SecretKey     string            `yaml:"secret_key"`      // 密钥（可选，如果启用了认证）
	RetryTimes    int               `yaml:"retry_times"`     // 发送重试次数（默认值）
	SendMsgTimeout int              `yaml:"send_msg_timeout"` // 发送消息超时（秒，默认值）
	EnableTrace   bool             `yaml:"enable_trace"`     // 是否开启消息轨迹
	ServiceGroups map[string]string `yaml:"service_groups"`  // 服务名到生产者组名的映射
	
	// 兼容 4.x 配置（如果配置了 name_servers，则使用 4.x 客户端）
	NameServers   []string          `yaml:"name_servers"`     // NameServer 地址列表（4.x 使用，5.x 不使用）
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
	ProductServiceAddr string `yaml:"product_service_addr"` // 商品服务 gRPC 地址，例如 "localhost:50053"
}

type Config struct {
	Services         map[string]ServiceConfig `yaml:"services"`
	MySQL            DatabaseConfig           `yaml:"mysql"`
	ServiceDatabases map[string]string        `yaml:"service_databases"` // 服务名到数据库名的映射
	Redis            RedisConfig              `yaml:"redis"`
	RocketMQ         RocketMQConfig           `yaml:"rocketmq"`
	SMS              SMSConfig                `yaml:"sms"`
	JWT              JWTConfig                `yaml:"jwt"`
	OSS              OSSConfig                `yaml:"oss"`
	Elasticsearch    ElasticsearchConfig      `yaml:"elasticsearch"`
	ServiceClients   ServiceClientConfig      `yaml:"service_clients"` // 服务客户端配置
}

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

// GetRocketMQConfig 获取 RocketMQ 配置（全局公共配置）
func (c *Config) GetRocketMQConfig() *RocketMQConfig {
	return &c.RocketMQ
}

// GetRocketMQConfigForService 获取指定服务的 RocketMQ 配置
// 返回该服务的生产者组名和公共配置
func (c *Config) GetRocketMQConfigForService(serviceName string) (groupName string, config *RocketMQConfig, err error) {
	config = &c.RocketMQ

	// 从服务映射中获取该服务的生产者组名
	if c.RocketMQ.ServiceGroups != nil {
		if gName, exists := c.RocketMQ.ServiceGroups[serviceName]; exists {
			groupName = gName
		}
	}

	// 如果没有配置，使用默认命名约定：service-name -> service-name-producer-group
	if groupName == "" {
		groupName = serviceName + "-producer-group"
	}

	return groupName, config, nil
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
