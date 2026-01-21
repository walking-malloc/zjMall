package registry

import (
	"fmt"
	"log"

	"zjMall/internal/config"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

func NewNacosNamingClient(cfg *config.NacosConfig) (naming_client.INamingClient, error) {
	sc := []constant.ServerConfig{
		*constant.NewServerConfig(cfg.Host, cfg.Port),
	}
	cc := *constant.NewClientConfig(
		constant.WithNamespaceId(cfg.Namespace),
		constant.WithUsername(cfg.Username),
		constant.WithPassword(cfg.Password),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogLevel("warn"),
	)
	return clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
}

func RegisterService(nc naming_client.INamingClient, serviceName, ip string, port uint64) error {
	success, err := nc.RegisterInstance(vo.RegisterInstanceParam{
		ServiceName: serviceName,
		Ip:          ip,
		Port:        port,
		Weight:      1.0,
		Healthy:     true,
		Enable:      true,
		Ephemeral:   true, // 临时实例，断开自动下线
	})
	if err != nil {
		return err
	}
	log.Printf("Nacos register %s %s:%d => %v", serviceName, ip, port, success)
	return nil
}

// SelectOneHealthyInstance 从 Nacos 选择一个健康实例，返回 "ip:port"
func SelectOneHealthyInstance(nc naming_client.INamingClient, serviceName string) (string, error) {
	inst, err := nc.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: serviceName,
	})
	if err != nil {
		return "", err
	}
	if inst == nil {
		return "", fmt.Errorf("no healthy instance for service %s", serviceName)
	}
	addr := fmt.Sprintf("%s:%d", inst.Ip, inst.Port)
	return addr, nil
}
