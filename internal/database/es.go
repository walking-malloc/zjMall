package database

import (
	"context"
	"fmt"
	"log"
	"zjMall/internal/config"

	"github.com/elastic/go-elasticsearch/v8"
)

type ElasticsearchClient struct {
	client *elasticsearch.Client
}

func NewElasticsearchClient(config *config.ElasticsearchConfig) (*ElasticsearchClient, error) {
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{config.Host},
	})
	if err != nil {
		return nil, fmt.Errorf("初始化 Elasticsearch 客户端失败: %v", err)
	}

	// 测试连接
	res, err := es.Info()
	if err != nil {
		return nil, fmt.Errorf("ES连接失败: %w", err)
	}
	defer res.Body.Close()

	log.Println("✅ Elasticsearch 连接成功")

	return &ElasticsearchClient{client: es}, nil
}

func (e *ElasticsearchClient) GetClient() *elasticsearch.Client {
	return e.client
}

func (e *ElasticsearchClient) Close(ctx context.Context) error {
	return e.client.Close(ctx)
}
