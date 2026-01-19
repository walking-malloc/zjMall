package database

import (
	"fmt"
	"log"
	"zjMall/internal/config"

	rmq "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
)

// RocketMQProducer 全局生产者实例（5.x gRPC 客户端接口类型）
var RocketMQProducer rmq.Producer

// InitRocketMQ 初始化 RocketMQ 5.x 生产者
// groupName: 生产者组名（每个服务应该有不同的组名）
// cfg: RocketMQ 公共配置（Endpoint、重试次数等）
func InitRocketMQ(groupName string, cfg *config.RocketMQConfig) (rmq.Producer, error) {
	// 检查是否配置了 Endpoint（5.x）或 NameServers（4.x）
	if cfg.Endpoint == "" && len(cfg.NameServers) == 0 {
		return nil, fmt.Errorf("RocketMQ 配置错误：必须配置 endpoint（5.x）或 name_servers（4.x）")
	}

	// 如果配置了 NameServers，说明还在用 4.x，返回错误提示升级
	if len(cfg.NameServers) > 0 && cfg.Endpoint == "" {
		return nil, fmt.Errorf("RocketMQ 4.x 客户端已弃用，请升级到 5.x 并配置 endpoint（例如：127.0.0.1:8081）")
	}

	// RocketMQ 5.x 使用 gRPC Proxy Endpoint
	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = "127.0.0.1:8081" // 默认 Proxy gRPC 端口
	}

	// 构建配置
	producerConfig := &rmq.Config{
		Endpoint: endpoint,
	}

	// 为避免 RocketMQ 5.x SDK 在 Credentials 为空指针时 Sign 发生 nil panic，
	// 始终提供一个非空的 SessionCredentials（即使不启用 ACL 也可以是空值）。
	producerConfig.Credentials = &credentials.SessionCredentials{
		AccessKey:    cfg.AccessKey,
		AccessSecret: cfg.SecretKey,
	}

	// TODO: 如果后续需要，可以在这里使用 cfg.RetryTimes、cfg.SendMsgTimeout、cfg.EnableTrace
	// 来构建重试策略、超时和消息轨迹配置。目前 Go v5 客户端主要通过默认策略工作。

	// 创建生产者（注意：返回的是接口类型 rmq.Producer）
	p, err := rmq.NewProducer(producerConfig, rmq.WithTopics("cart-sync")) // 预订阅 Topic
	if err != nil {
		return nil, fmt.Errorf("创建 RocketMQ 5.x 生产者失败: %w", err)
	}

	// 启动生产者
	if err := p.Start(); err != nil {
		return nil, fmt.Errorf("启动 RocketMQ 5.x 生产者失败: %w", err)
	}

	log.Printf("✅ RocketMQ 5.x 生产者启动成功: GroupName=%s, Endpoint=%s", groupName, endpoint)
	RocketMQProducer = p
	return p, nil
}

// CloseRocketMQ 关闭 RocketMQ 生产者
func CloseRocketMQ() error {
	if RocketMQProducer == nil {
		return nil
	}
	return RocketMQProducer.GracefulStop()
}

// GetRocketMQProducer 获取 RocketMQ 生产者实例
func GetRocketMQProducer() rmq.Producer {
	return RocketMQProducer
}
