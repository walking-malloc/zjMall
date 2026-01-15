package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
	"zjMall/internal/product-service/model"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

const (
	ProductIndexName = "products"
)

type SearchRepository interface {
	// 索引操作
	CreateIndex(ctx context.Context) error
	IndexProduct(ctx context.Context, product *model.ProductIndex) error
	BulkIndexProducts(ctx context.Context, products []*model.ProductIndex) error
	DeleteProduct(ctx context.Context, productID string) error
	UpdateProduct(ctx context.Context, product *model.ProductIndex) error

	// 搜索操作
	SearchProducts(ctx context.Context, keyword string, page, pageSize int32, filters *SearchFilters) (*SearchResult, error)
}

type SearchFilters struct {
	CategoryID string
	BrandID    string
	Status     int8
	MinPrice   float64
	MaxPrice   float64
	Tags       []string
}

type SearchResult struct {
	Total    int64
	Products []*model.ProductIndex
}

type searchRepository struct {
	esClient *elasticsearch.Client
}

func NewSearchRepository(esClient *elasticsearch.Client) SearchRepository {
	return &searchRepository{
		esClient: esClient,
	}
}

// CreateIndex 创建商品索引
func (r *searchRepository) CreateIndex(ctx context.Context) error {
	// 先检查索引是否已存在
	existsReq := esapi.IndicesExistsRequest{
		Index: []string{ProductIndexName},
	}
	existsRes, err := existsReq.Do(ctx, r.esClient)
	if err == nil && existsRes != nil {
		existsRes.Body.Close()
		if existsRes.StatusCode == 200 {
			log.Printf("⚠️  索引 %s 已存在，删除旧索引以应用新的映射...", ProductIndexName)
			// 删除旧索引
			deleteReq := esapi.IndicesDeleteRequest{
				Index: []string{ProductIndexName},
			}
			deleteRes, err := deleteReq.Do(ctx, r.esClient)
			if err != nil {
				log.Printf("⚠️  删除旧索引失败: %v，继续创建新索引", err)
			} else if deleteRes != nil {
				deleteRes.Body.Close()
				log.Printf("✅ 旧索引已删除")
			}
		}
	}

	// 使用简化的索引配置（去掉不必要的 analyzer 配置）
	indexBody := `
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 1
  },
  "mappings": {
    "properties": {
      "id": {
        "type": "keyword"
      },
      "title": {
        "type": "text",
        "analyzer": "standard",
        "search_analyzer": "standard",
        "fields": {
          "keyword": {
            "type": "keyword"
          }
        }
      },
      "subtitle": {
        "type": "text",
        "analyzer": "standard",
        "search_analyzer": "standard"
      },
      "description": {
        "type": "text",
        "analyzer": "standard",
        "search_analyzer": "standard"
      },
      "category_id": {
        "type": "keyword"
      },
      "category_name": {
        "type": "text",
        "analyzer": "standard"
      },
      "brand_id": {
        "type": "keyword"
      },
      "brand_name": {
        "type": "text",
        "analyzer": "standard"
      },
      "tags": {
        "type": "keyword"
      },
      "skus": {
        "type": "nested",
        "properties": {
          "sku_name": {
            "type": "keyword"
          },
          "price": {
            "type": "float"
          }
        }
      },
      "attribute_values": {
        "type": "keyword"
      },
      "attributes": {
        "type": "nested",
        "properties": {
          "attribute_id": {
            "type": "keyword"
          },
          "attribute_name": {
            "type": "text",
            "analyzer": "standard"
          },
          "value": {
            "type": "text",
            "analyzer": "standard"
          }
        }
      },
      "status": {
        "type": "byte"
      },
      "on_shelf_time": {
        "type": "date",
        "format": "strict_date_optional_time||epoch_millis"
      },
      "created_at": {
        "type": "date",
        "format": "strict_date_optional_time||epoch_millis"
      },
      "updated_at": {
        "type": "date",
        "format": "strict_date_optional_time||epoch_millis"
      }
    }
  }
}`

	req := esapi.IndicesCreateRequest{
		Index: ProductIndexName,
		Body:  strings.NewReader(indexBody),
	}

	res, err := req.Do(ctx, r.esClient)
	if err != nil {
		log.Printf("❌ 创建索引请求失败: %v", err)
		return fmt.Errorf("创建索引失败: %w", err)
	}

	// 确保响应体被关闭
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}

	// 读取响应体
	var responseBody bytes.Buffer
	if res.Body != nil {
		_, err := responseBody.ReadFrom(res.Body)
		if err != nil {
			log.Printf("⚠️  读取响应体失败: %v", err)
		}
	}

	log.Printf("📊 创建索引响应状态码: %d", res.StatusCode)
	log.Printf("📊 创建索引响应内容: %s", responseBody.String())

	if res.IsError() {
		// 读取错误响应
		errorMsg := responseBody.String()

		// 如果索引已存在，忽略错误
		if res.StatusCode == 400 {
			if strings.Contains(errorMsg, "already exists") ||
				strings.Contains(errorMsg, "resource_already_exists_exception") ||
				strings.Contains(errorMsg, "index_already_exists_exception") {
				log.Printf("✅ 索引 %s 已存在（从错误响应中检测到）", ProductIndexName)
				return nil
			}
			// 其他 400 错误，返回详细信息
			log.Printf("❌ 创建索引错误 [400]: %s", errorMsg)
			return fmt.Errorf("创建索引错误 [400]: %s", errorMsg)
		}
		log.Printf("❌ 创建索引错误 [%d]: %s", res.StatusCode, errorMsg)
		return fmt.Errorf("创建索引错误 [%d]: %s", res.StatusCode, errorMsg)
	}

	log.Printf("✅ 索引 %s 创建成功", ProductIndexName)
	return nil
}

// fixDateTimeFormat 修复日期时间格式为 RFC3339
func fixDateTimeFormat(dateStr string) (string, error) {
	if dateStr == "" {
		return "", fmt.Errorf("日期字符串为空")
	}

	// 如果已经是 RFC3339 格式，直接返回
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return t.Format(time.RFC3339), nil
	}

	// 尝试多种时间格式（按常见程度排序）
	formats := []string{
		"2006-01-02 15:04:05",           // MySQL 默认格式
		"2006-01-02T15:04:05Z07:00",     // RFC3339 变体
		"2006-01-02T15:04:05Z",          // UTC 格式
		"2006-01-02T15:04:05",           // 无时区格式
		"2006-01-02 15:04:05.000000",    // MySQL 微秒格式
		"2006-01-02 15:04:05.000000000", // MySQL 纳秒格式
		time.RFC3339Nano,                // RFC3339 纳秒格式
		"2006-01-02",                    // 仅日期
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			// 使用本地时区格式化
			return t.Format(time.RFC3339), nil
		}
	}

	// 如果所有格式都失败，尝试使用 time.ParseInLocation（使用本地时区）
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", dateStr, time.Local); err == nil {
		return t.Format(time.RFC3339), nil
	}

	return "", fmt.Errorf("无法解析时间格式: %s", dateStr)
}

// IndexProduct 索引单个商品
func (r *searchRepository) IndexProduct(ctx context.Context, product *model.ProductIndex) error {
	// 修复所有日期格式为 RFC3339
	if product.OnShelfTime != nil {
		fixed, err := fixDateTimeFormat(*product.OnShelfTime)
		if err != nil {
			return fmt.Errorf("OnShelfTime 格式修复失败: %s, 错误: %w", *product.OnShelfTime, err)
		}
		product.OnShelfTime = &fixed
	}

	if fixed, err := fixDateTimeFormat(product.CreatedAt); err == nil {
		product.CreatedAt = fixed
	}

	if fixed, err := fixDateTimeFormat(product.UpdatedAt); err == nil {
		product.UpdatedAt = fixed
	}

	body, err := json.Marshal(product)
	if err != nil {
		return fmt.Errorf("序列化商品失败: %w", err)
	}

	// 最终验证：检查 JSON 中的日期格式，如果不对就强制修复
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(body, &jsonMap); err == nil {
		needRemarshal := false

		if onShelfTime, ok := jsonMap["on_shelf_time"].(string); ok {
			if _, err := time.Parse(time.RFC3339, onShelfTime); err != nil {
				fixed, fixErr := fixDateTimeFormat(onShelfTime)
				if fixErr == nil {
					jsonMap["on_shelf_time"] = fixed
					needRemarshal = true
				} else {
					return fmt.Errorf("JSON 中的 on_shelf_time 格式不正确且无法修复: %s", onShelfTime)
				}
			}
		}

		if createdAt, ok := jsonMap["created_at"].(string); ok {
			if _, err := time.Parse(time.RFC3339, createdAt); err != nil {
				if fixed, fixErr := fixDateTimeFormat(createdAt); fixErr == nil {
					jsonMap["created_at"] = fixed
					needRemarshal = true
				}
			}
		}

		if updatedAt, ok := jsonMap["updated_at"].(string); ok {
			if _, err := time.Parse(time.RFC3339, updatedAt); err != nil {
				if fixed, fixErr := fixDateTimeFormat(updatedAt); fixErr == nil {
					jsonMap["updated_at"] = fixed
					needRemarshal = true
				}
			}
		}

		if needRemarshal {
			body, err = json.Marshal(jsonMap)
			if err != nil {
				return fmt.Errorf("重新序列化失败: %w", err)
			}
		}
	}

	req := esapi.IndexRequest{
		Index:      ProductIndexName,
		DocumentID: product.ID,
		Body:       bytes.NewReader(body),
		Refresh:    "true", // 写入后立即刷新，使文档可搜索
	}

	res, err := req.Do(ctx, r.esClient)
	if err != nil {
		return fmt.Errorf("索引商品失败: %w", err)
	}
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}

	if res.IsError() {
		// 读取错误响应体以获取详细信息
		var errorBody bytes.Buffer
		if res.Body != nil {
			errorBody.ReadFrom(res.Body)
		}
		return fmt.Errorf("索引商品错误 [%d]: %s", res.StatusCode, errorBody.String())
	}

	return nil
}

// BulkIndexProducts 批量索引商品
func (r *searchRepository) BulkIndexProducts(ctx context.Context, products []*model.ProductIndex) error {
	var buf bytes.Buffer

	for _, product := range products {
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": ProductIndexName,
				"_id":    product.ID,
			},
		}

		if err := json.NewEncoder(&buf).Encode(meta); err != nil {
			return fmt.Errorf("编码元数据失败: %w", err)
		}

		if err := json.NewEncoder(&buf).Encode(product); err != nil {
			return fmt.Errorf("编码商品失败: %w", err)
		}
	}

	res, err := r.esClient.Bulk(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return fmt.Errorf("批量索引失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("批量索引错误: %s", res.String())
	}

	return nil
}

// DeleteProduct 删除商品索引
func (r *searchRepository) DeleteProduct(ctx context.Context, productID string) error {
	req := esapi.DeleteRequest{
		Index:      ProductIndexName,
		DocumentID: productID,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, r.esClient)
	if err != nil {
		return fmt.Errorf("删除商品索引失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("删除商品索引错误: %s", res.String())
	}

	return nil
}

// UpdateProduct 更新商品索引
func (r *searchRepository) UpdateProduct(ctx context.Context, product *model.ProductIndex) error {
	return r.IndexProduct(ctx, product) // ES 的更新就是重新索引
}

// SearchProducts 搜索商品
func (r *searchRepository) SearchProducts(ctx context.Context, keyword string, page, pageSize int32, filters *SearchFilters) (*SearchResult, error) {
	var query map[string]interface{}

	if keyword == "" {
		// 无关键词，使用 match_all
		query = map[string]interface{}{
			"match_all": map[string]interface{}{},
		}
	} else {
		// 多字段搜索
		query = map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{ //满足一个条件即可
					{
						"match": map[string]interface{}{
							"title": map[string]interface{}{
								"query": keyword,
								"boost": 3.0, // 标题权重最高
							},
						},
					},
					{
						"match": map[string]interface{}{
							"subtitle": map[string]interface{}{
								"query": keyword,
								"boost": 2.0,
							},
						},
					},
					{
						"match": map[string]interface{}{
							"description": map[string]interface{}{
								"query": keyword,
								"boost": 1.0,
							},
						},
					},
					{
						"match": map[string]interface{}{
							"category_name": map[string]interface{}{
								"query": keyword,
								"boost": 1.5,
							},
						},
					},
					{
						"match": map[string]interface{}{
							"brand_name": map[string]interface{}{
								"query": keyword,
								"boost": 1.5,
							},
						},
					},
					{
						"match": map[string]interface{}{
							"attribute_values": map[string]interface{}{
								"query": keyword,
								"boost": 1.0,
							},
						},
					},
					{
						"nested": map[string]interface{}{ //处理嵌套结构
							"path": "skus",
							"query": map[string]interface{}{
								"bool": map[string]interface{}{
									"should": []map[string]interface{}{
										{
											"match": map[string]interface{}{
												"skus.sku_name": map[string]interface{}{
													"query": keyword,
													"boost": 1.2,
												},
											},
										},
									},
								},
							},
						},
					},
				},
				"minimum_should_match": 1, //至少匹配一个should条件
			},
		}
	}

	// 构建过滤条件
	must := []map[string]interface{}{query}

	if filters != nil {
		// 状态过滤（只搜索已上架商品）
		if filters.Status > 0 {
			must = append(must, map[string]interface{}{
				"term": map[string]interface{}{
					"status": filters.Status,
				},
			})
		}

		// 类目过滤
		if filters.CategoryID != "" {
			must = append(must, map[string]interface{}{
				"term": map[string]interface{}{
					"category_id": filters.CategoryID,
				},
			})
		}

		// 品牌过滤
		if filters.BrandID != "" {
			must = append(must, map[string]interface{}{
				"term": map[string]interface{}{
					"brand_id": filters.BrandID,
				},
			})
		}

		// 标签过滤
		if len(filters.Tags) > 0 {
			must = append(must, map[string]interface{}{
				"terms": map[string]interface{}{
					"tags": filters.Tags,
				},
			})
		}
	}

	// 构建完整查询
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": must,
			},
		},
		"from": (page - 1) * pageSize,
		"size": pageSize,
		"sort": []map[string]interface{}{ //排序
			{"_score": map[string]interface{}{"order": "desc"}},        //按相关性分数排序
			{"on_shelf_time": map[string]interface{}{"order": "desc"}}, //按上架时间排序
		},
	}

	body, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, fmt.Errorf("序列化查询失败: %w", err)
	}

	res, err := r.esClient.Search(
		r.esClient.Search.WithContext(ctx),
		r.esClient.Search.WithIndex(ProductIndexName),
		r.esClient.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, fmt.Errorf("搜索失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		// 读取错误响应体
		var errorBody bytes.Buffer
		if res.Body != nil {
			errorBody.ReadFrom(res.Body)
		}
		log.Printf("❌ ES搜索错误 [%d]: %s", res.StatusCode, errorBody.String())
		return nil, fmt.Errorf("搜索错误 [%d]: %s", res.StatusCode, errorBody.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析搜索结果失败: %w", err)
	}
	log.Println("ES搜索返回结果:", result)
	hits := result["hits"].(map[string]interface{})
	total := int64(hits["total"].(map[string]interface{})["value"].(float64))
	hitsArray := hits["hits"].([]interface{})

	products := make([]*model.ProductIndex, 0, len(hitsArray))
	for _, hit := range hitsArray {
		hitMap := hit.(map[string]interface{})
		source := hitMap["_source"].(map[string]interface{})

		product := &model.ProductIndex{}

		// 基础字段
		if id, ok := source["id"].(string); ok {
			product.ID = id
		}
		if title, ok := source["title"].(string); ok {
			product.Title = title
		}
		if subtitle, ok := source["subtitle"].(string); ok {
			product.Subtitle = subtitle
		}
		if description, ok := source["description"].(string); ok {
			product.Description = description
		}
		if categoryID, ok := source["category_id"].(string); ok {
			product.CategoryID = categoryID
		}
		if categoryName, ok := source["category_name"].(string); ok {
			product.CategoryName = categoryName
		}
		if brandID, ok := source["brand_id"].(string); ok {
			product.BrandID = brandID
		}
		if brandName, ok := source["brand_name"].(string); ok {
			product.BrandName = brandName
		}

		// 状态字段（需要类型转换）
		if statusVal, ok := source["status"]; ok {
			switch v := statusVal.(type) {
			case float64:
				product.Status = int8(v)
			case int8:
				product.Status = v
			case int:
				product.Status = int8(v)
			}
		}

		// 标签数组（需要处理 interface{} 数组）
		if tagsVal, ok := source["tags"]; ok {
			if tagsArray, ok := tagsVal.([]interface{}); ok {
				tags := make([]string, 0, len(tagsArray))
				for _, tag := range tagsArray {
					if tagStr, ok := tag.(string); ok {
						tags = append(tags, tagStr)
					}
				}
				product.Tags = tags
			}
		}

		// SKU 嵌套数组
		if skusVal, ok := source["skus"]; ok {
			if skusArray, ok := skusVal.([]interface{}); ok {
				skus := make([]*model.SKUIndex, 0, len(skusArray))
				for _, skuVal := range skusArray {
					if skuMap, ok := skuVal.(map[string]interface{}); ok {
						sku := &model.SKUIndex{}
						if skuName, ok := skuMap["sku_name"].(string); ok {
							sku.SKUName = skuName
						}
						if priceVal, ok := skuMap["price"]; ok {
							switch v := priceVal.(type) {
							case float64:
								sku.Price = v
							case float32:
								sku.Price = float64(v)
							}
						}
						skus = append(skus, sku)
					}
				}
				product.SKUs = skus
			}
		}

		// 属性值数组
		if attributeValuesVal, ok := source["attribute_values"]; ok {
			if attrArray, ok := attributeValuesVal.([]interface{}); ok {
				attributeValues := make([]string, 0, len(attrArray))
				for _, attr := range attrArray {
					if attrStr, ok := attr.(string); ok {
						attributeValues = append(attributeValues, attrStr)
					}
				}
				product.AttributeValues = attributeValues
			}
		}

		// 时间字段
		if onShelfTime, ok := source["on_shelf_time"].(string); ok {
			product.OnShelfTime = &onShelfTime
		}
		if createdAt, ok := source["created_at"].(string); ok {
			product.CreatedAt = createdAt
		}
		if updatedAt, ok := source["updated_at"].(string); ok {
			product.UpdatedAt = updatedAt
		}

		products = append(products, product)
	}

	return &SearchResult{
		Total:    total,
		Products: products,
	}, nil
}
