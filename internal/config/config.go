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
type Config struct {
	Services         map[string]ServiceConfig `yaml:"services"`
	MySQL            DatabaseConfig           `yaml:"mysql"`
	ServiceDatabases map[string]string        `yaml:"service_databases"` // 服务名到数据库名的映射
	Redis            RedisConfig              `yaml:"redis"`
	SMS              SMSConfig                `yaml:"sms"`
	JWT              JWTConfig                `yaml:"jwt"`
	OSS              OSSConfig                `yaml:"oss"`
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

func (c *Config) GetMySQLConfig() *DatabaseConfig {
	return &c.MySQL
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

// GetDatabaseConfigForService 根据服务名返回对应的数据库配置
// 优先从配置文件中的 service_databases 映射获取，如果没有则使用命名约定
func (c *Config) GetDatabaseConfigForService(serviceName string) (*DatabaseConfig, error) {
	// 验证服务是否存在
	_, exists := c.Services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found in config", serviceName)
	}

	var dbName string

	// 优先从配置文件中获取映射
	if c.ServiceDatabases != nil {
		if mappedDbName, ok := c.ServiceDatabases[serviceName]; ok {
			dbName = mappedDbName
		}
	}

	// 如果配置文件中没有映射，使用命名约定：service-name -> service_db
	if dbName == "" {
		// 去掉 "-service" 后缀，然后加上 "_db"
		// 例如：user-service -> user_db
		if len(serviceName) > 8 && serviceName[len(serviceName)-8:] == "-service" {
			dbName = serviceName[:len(serviceName)-8] + "_db"
		} else {
			// 如果没有 "-service" 后缀，直接加上 "_db"
			dbName = serviceName + "_db"
		}
	}

	// 复制基础 MySQL 配置
	dbConfig := c.MySQL
	// 设置服务特定的数据库名
	dbConfig.DBName = dbName

	return &dbConfig, nil
}
