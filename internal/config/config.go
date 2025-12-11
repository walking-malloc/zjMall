package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type ServiceConfig struct {
	GRPC struct {
		Port int `yaml:"port"`
	} `yaml:"grpc"`
	HTTP struct {
		Port int `yaml:"port"`
	} `yaml:"http"`
}
type Config struct {
	Services map[string]ServiceConfig `yaml:"services"`
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
