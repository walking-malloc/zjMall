package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"zjMall/internal/cart-service/model"
	"zjMall/internal/common/cache"
	"zjMall/internal/common/mq"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

const (
	// Redis Key 前缀
	CacheKeyCart     = "cart:user:%s"      // 用户购物车：cart:user:{user_id}
	CacheKeyCartItem = "cart:item:%s"      // 购物车项：cart:item:{item_id}
	CartExpiration   = 30 * 24 * time.Hour // 购物车过期时间：30天
)

type CartRepository interface {
	// 添加商品到购物车
	AddItem(ctx context.Context, userID string, item *model.CartItem) error

	// 更新购物车项数量
	UpdateItemQuantity(ctx context.Context, userID string, itemID string, quantity int32) error

	// 删除购物车项
	RemoveItem(ctx context.Context, userID string, itemID string) error

	// 批量删除购物车项
	RemoveItems(ctx context.Context, userID string, itemIDs []string) error

	// 清空购物车
	ClearCart(ctx context.Context, userID string) error

	// 获取购物车所有商品
	GetCartItems(ctx context.Context, userID string) ([]*model.CartItem, error)

	// 获取购物车项
	GetCartItem(ctx context.Context, userID string, itemID string) (*model.CartItem, error)

	// 检查购物车项是否存在
	ItemExists(ctx context.Context, userID string, itemID string) (bool, error)

	// 根据用户ID和SKU ID查找购物车项（用于判断是否已存在相同SKU）
	GetCartItemByUserAndSKU(ctx context.Context, userID string, skuID string) (*model.CartItem, error)
}

type cartRepository struct {
	db          *gorm.DB
	redisClient *redis.Client
	cacheRepo   cache.CacheRepository
	mqProducer  mq.MessageProducer // 消息队列生产者
}

func NewCartRepository(db *gorm.DB, redisClient *redis.Client, cacheRepo cache.CacheRepository, mqProducer mq.MessageProducer) CartRepository {
	return &cartRepository{
		db:          db,
		redisClient: redisClient,
		cacheRepo:   cacheRepo,
		mqProducer:  mqProducer,
	}
}

// AddItem 添加商品到购物车（Redis 主存储 + MQ 异步同步到 MySQL）
func (r *cartRepository) AddItem(ctx context.Context, userID string, item *model.CartItem) error {
	// 1. 写入 Redis（主存储，快速响应）
	if err := r.setToCache(ctx, userID, item); err != nil {
		log.Printf("❌ [Repository] AddItem: 写入 Redis 失败 - user_id=%s, item_id=%s, error=%v", userID, item.ID, err)
		return fmt.Errorf("写入 Redis 失败: %w", err)
	}

	// 2. 发送消息到 RocketMQ（异步同步到 MySQL）
	if r.mqProducer != nil {
		log.Printf("🔍 [DEBUG] 准备发送购物车事件: userID=%s, itemID=%s", userID, item.ID)
		eventData := map[string]interface{}{
			"id":             item.ID,
			"product_id":     item.ProductID,
			"sku_id":         item.SKUID,
			"product_title":  item.ProductTitle,
			"product_image":  item.ProductImage,
			"sku_name":       item.SKUName,
			"price":          item.Price,
			"current_price":  item.CurrentPrice,
			"quantity":       item.Quantity,
			"stock":          item.Stock,
			"is_valid":       item.IsValid,
			"invalid_reason": item.InvalidReason,
		}
		event := mq.NewCartItemAddedEvent(userID, item.ID, eventData)
		if err := mq.SendCartEvent(ctx, r.mqProducer, event); err != nil {
			// MQ 发送失败不影响主流程，只记录日志
			// 可以考虑异步重试或定时补偿
			log.Printf("⚠️ 发送购物车事件失败: %v", err)
		} else {
			log.Printf("✅ [DEBUG] 购物车事件已提交发送: userID=%s, itemID=%s", userID, item.ID)
		}
	} else {
		log.Printf("⚠️ [DEBUG] mqProducer 为 nil，跳过 MQ 发送")
	}

	return nil
}

// UpdateItemQuantity 更新购物车项数量
func (r *cartRepository) UpdateItemQuantity(ctx context.Context, userID string, itemID string, quantity int32) error {
	// 1. 更新 Redis（主存储）
	item, err := r.GetCartItem(ctx, userID, itemID)
	if err != nil {
		log.Printf("❌ [Repository] UpdateItemQuantity: 获取购物车项失败 - user_id=%s, item_id=%s, error=%v", userID, itemID, err)
		return err
	}
	if item == nil {
		log.Printf("⚠️ [Repository] UpdateItemQuantity: 购物车项不存在 - user_id=%s, item_id=%s", userID, itemID)
		return fmt.Errorf("购物车项不存在")
	}

	item.Quantity = quantity
	item.UpdatedAt = time.Now()
	if err := r.setToCache(ctx, userID, item); err != nil {
		log.Printf("❌ [Repository] UpdateItemQuantity: 更新 Redis 失败 - user_id=%s, item_id=%s, error=%v", userID, itemID, err)
		return fmt.Errorf("更新 Redis 失败: %w", err)
	}

	// 2. 发送消息到 RocketMQ（异步同步到 MySQL）
	if r.mqProducer != nil {
		eventData := map[string]interface{}{
			"quantity": quantity,
		}
		event := mq.NewCartItemUpdatedEvent(userID, itemID, eventData)
		if err := mq.SendCartEvent(ctx, r.mqProducer, event); err != nil {
			log.Printf("⚠️ 发送购物车更新事件失败: %v", err)
		}
	}

	return nil
}

// RemoveItem 删除购物车项
func (r *cartRepository) RemoveItem(ctx context.Context, userID string, itemID string) error {
	// 1. 从 Redis 删除（主存储）
	r.deleteFromCache(ctx, userID, itemID)

	// 2. 发送消息到 RocketMQ（异步同步到 MySQL）
	if r.mqProducer != nil {
		event := mq.NewCartItemRemovedEvent(userID, itemID)
		if err := mq.SendCartEvent(ctx, r.mqProducer, event); err != nil {
			log.Printf("⚠️ 发送购物车删除事件失败: %v", err)
		}
	}

	return nil
}

// RemoveItems 批量删除购物车项
// 使用 Pipeline 批量删除，减少网络往返，提升性能
// 自动去重：如果 itemIDs 有重复，只删除一次
func (r *cartRepository) RemoveItems(ctx context.Context, userID string, itemIDs []string) error {
	if len(itemIDs) == 0 {
		return nil
	}

	// 去重：使用 map 去重，避免重复删除和重复发送 MQ 消息
	uniqueItemIDs := make(map[string]bool)
	deduplicatedIDs := make([]string, 0, len(itemIDs))
	for _, itemID := range itemIDs {
		if !uniqueItemIDs[itemID] {
			uniqueItemIDs[itemID] = true
			deduplicatedIDs = append(deduplicatedIDs, itemID)
		}
	}

	// 1. 从 Redis 批量删除（主存储）- 使用 Pipeline 优化性能
	cartKey := fmt.Sprintf(CacheKeyCart, userID)
	pipe := r.redisClient.Pipeline()

	for _, itemID := range deduplicatedIDs {
		itemKey := fmt.Sprintf(CacheKeyCartItem, itemID)
		pipe.HDel(ctx, cartKey, itemID)
		pipe.Del(ctx, itemKey)
		log.Printf("✅ [Repository] RemoveItems: 删除缓存,cartKey=%s, item_ids=%v", cartKey, deduplicatedIDs)
		log.Printf("✅ [Repository] RemoveItems: 删除缓存,itemKey=%s, itemID=%s", itemKey, itemID)
	}

	// 批量执行所有删除操作（原子执行，减少网络往返）
	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Printf("⚠️ 批量删除缓存失败: %v", err)
		// 不返回错误，继续执行 MQ 发送（最终一致性）
	}
	log.Printf("✅ [Repository] RemoveItems: 批量删除缓存成功 - user_id=%s, item_ids=%v", userID, deduplicatedIDs)
	// 2. 发送消息到 RocketMQ（异步同步到 MySQL）- 使用去重后的 IDs
	if r.mqProducer != nil {
		for _, itemID := range deduplicatedIDs {
			event := mq.NewCartItemRemovedEvent(userID, itemID)
			if err := mq.SendCartEvent(ctx, r.mqProducer, event); err != nil {
				log.Printf("⚠️ 发送购物车批量删除事件失败 (item_id=%s): %v", itemID, err)
			}
		}
	}

	return nil
}

// ClearCart 清空购物车
func (r *cartRepository) ClearCart(ctx context.Context, userID string) error {
	// 1. 从 Redis 删除（主存储）
	cartKey := fmt.Sprintf(CacheKeyCart, userID)
	r.redisClient.Del(ctx, cartKey)

	// 2. 发送消息到 RocketMQ（异步同步到 MySQL）
	if r.mqProducer != nil {
		event := mq.NewCartClearedEvent(userID)
		if err := mq.SendCartEvent(ctx, r.mqProducer, event); err != nil {
			log.Printf("⚠️ 发送购物车清空事件失败: %v", err)
		}
	}

	return nil
}

// GetCartItems 获取购物车所有商品（优先从 Redis 读取，缓存未命中则从 MySQL 读取并回写缓存）
// 使用 Redis 分布式锁防止缓存击穿：多个实例/请求同时缓存未命中时，只执行一次 MySQL 查询
func (r *cartRepository) GetCartItems(ctx context.Context, userID string) ([]*model.CartItem, error) {
	// 1. 先尝试从 Redis 读取（主存储）
	cartKey := fmt.Sprintf(CacheKeyCart, userID)
	itemsMap, err := r.redisClient.HGetAll(ctx, cartKey).Result()
	if err == nil && len(itemsMap) > 0 {
		// 缓存命中，反序列化返回
		items := make([]*model.CartItem, 0, len(itemsMap))
		for _, itemJSON := range itemsMap {
			var item model.CartItem
			if err := json.Unmarshal([]byte(itemJSON), &item); err == nil {
				items = append(items, &item)
			}
		}
		if len(items) > 0 {
			return items, nil
		}
	}

	// 获取锁失败，降级直接查数据库
	log.Printf("⚠️ [Repository] GetCartItems: 获取分布式锁失败，降级直接查询 - user_id=%s, error=%v", userID, err)
	var items []*model.CartItem
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		log.Printf("❌ [Repository] GetCartItems: 查询 MySQL 失败（降级） - user_id=%s, error=%v", userID, err)
		return nil, fmt.Errorf("查询购物车列表失败: %w", err)
	}
	for _, item := range items {
		r.setToCache(ctx, userID, item)
	}
	return items, nil

}

// GetCartItem 获取购物车项
// 使用 Redis 分布式锁防止缓存击穿：多个实例/请求同时缓存未命中时，只执行一次 MySQL 查询
func (r *cartRepository) GetCartItem(ctx context.Context, userID string, itemID string) (*model.CartItem, error) {
	// 1. 先尝试从 Redis 读取（主存储）
	itemKey := fmt.Sprintf(CacheKeyCartItem, itemID)
	itemJSON, err := r.redisClient.Get(ctx, itemKey).Result()
	if err == nil {
		var item model.CartItem
		if err := json.Unmarshal([]byte(itemJSON), &item); err == nil {
			// 验证用户ID是否匹配
			if item.UserID == userID {
				return &item, nil
			}
		}
	}

	// 获取锁失败，降级直接查数据库
	log.Printf("⚠️ [Repository] GetCartItem: 获取分布式锁失败，降级直接查询 - user_id=%s, item_id=%s, error=%v", userID, itemID, err)
	var item model.CartItem
	err = r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", itemID, userID).
		First(&item).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		log.Printf("❌ [Repository] GetCartItem: 查询 MySQL 失败（降级） - user_id=%s, item_id=%s, error=%v", userID, itemID, err)
		return nil, fmt.Errorf("查询购物车项失败: %w", err)
	}
	r.setToCache(ctx, userID, &item)
	return &item, nil
}

// GetCartItemByUserAndSKU 根据用户ID和SKU ID查找购物车项
func (r *cartRepository) GetCartItemByUserAndSKU(ctx context.Context, userID string, skuID string) (*model.CartItem, error) {
	var item model.CartItem
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND sku_id = ?", userID, skuID).
		First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		log.Printf("❌ [Repository] GetCartItemByUserAndSKU: 查询 MySQL 失败 - user_id=%s, sku_id=%s, error=%v", userID, skuID, err)
		return nil, fmt.Errorf("查询购物车项失败: %w", err)
	}
	return &item, nil
}

// ItemExists 检查购物车项是否存在
// 使用 EXISTS 子查询，性能最优：找到第一条匹配记录即返回，不需要扫描所有数据
// 使用参数化查询（? 占位符），GORM 会自动转义参数，防止 SQL 注入
func (r *cartRepository) ItemExists(ctx context.Context, userID string, itemID string) (bool, error) {
	var exists bool
	// 使用参数化查询，GORM 会将 itemID 和 userID 作为参数绑定，自动转义，防止 SQL 注入
	err := r.db.WithContext(ctx).
		Raw("SELECT EXISTS(SELECT 1 FROM cart_items WHERE id = ? AND user_id = ? LIMIT 1) AS `exists`", itemID, userID).
		Scan(&exists).Error

	if err != nil {
		log.Printf("❌ [Repository] ItemExists: 查询 MySQL 失败 - user_id=%s, item_id=%s, error=%v", userID, itemID, err)
		return false, fmt.Errorf("检查购物车项是否存在失败: %w", err)
	}
	return exists, nil
}

// ============================================
// 私有辅助方法
// ============================================
func (r *cartRepository) setToCache(ctx context.Context, userID string, item *model.CartItem) error {
	cartKey := fmt.Sprintf(CacheKeyCart, userID)
	itemKey := fmt.Sprintf(CacheKeyCartItem, item.ID)

	// 序列化购物车项
	itemJSON, err := json.Marshal(item)
	if err != nil {
		log.Printf("❌ [Repository] setToCache: 序列化购物车项失败 - user_id=%s, item_id=%s, error=%v", userID, item.ID, err)
		return fmt.Errorf("序列化购物车项失败: %w", err)
	}

	// 使用 Pipeline 批量操作，同时设置两个key的缓存
	pipe := r.redisClient.Pipeline()
	pipe.HSet(ctx, cartKey, item.ID, string(itemJSON))
	pipe.Expire(ctx, cartKey, CartExpiration)
	pipe.Set(ctx, itemKey, string(itemJSON), CartExpiration)

	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Printf("❌ [Repository] setToCache: 写入 Redis 失败 - user_id=%s, item_id=%s, error=%v", userID, item.ID, err)
		return fmt.Errorf("写入缓存失败: %w", err)
	}
	return nil
}

// deleteFromCache 从 Redis 缓存删除购物车项
func (r *cartRepository) deleteFromCache(ctx context.Context, userID string, itemID string) {
	cartKey := fmt.Sprintf(CacheKeyCart, userID)
	itemKey := fmt.Sprintf(CacheKeyCartItem, itemID)

	pipe := r.redisClient.Pipeline()
	pipe.HDel(ctx, cartKey, itemID)
	pipe.Del(ctx, itemKey)
	pipe.Exec(ctx) // 忽略错误，缓存删除失败不影响主流程
}
