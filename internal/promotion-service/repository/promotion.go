package repository

import (
	"context"
	"errors"
	"strings"
	"zjMall/internal/common/cache"
	"zjMall/internal/promotion-service/model"

	"gorm.io/gorm"
)

type PromotionListFilter struct {
	Page     int32
	PageSize int32
	Type     int8
	Status   int8
	Keyword  string
	Offset   int
	Limit    int
}

type PromotionRepository interface {
	CreatePromotion(ctx context.Context, promotion *model.Promotion) error
	GetPromotionByID(ctx context.Context, id string) (*model.Promotion, error)
	UpdatePromotion(ctx context.Context, promotion *model.Promotion) error
	DeletePromotion(ctx context.Context, id string) error
	ListPromotions(ctx context.Context, filter *PromotionListFilter) ([]*model.Promotion, int64, error)

	// 查询可用促销活动
	GetAvailablePromotions(ctx context.Context, productIDs []string, categoryID string, totalAmount float64) ([]*model.Promotion, error)

	// 记录促销使用
	RecordPromotionUsage(ctx context.Context, log *model.PromotionUsageLog) error
	GetUserPromotionUsageCount(ctx context.Context, promotionID, userID string) (int32, error)
}

type promotionRepository struct {
	db        *gorm.DB
	cacheRepo cache.CacheRepository
}

func NewPromotionRepository(db *gorm.DB, cacheRepo cache.CacheRepository) PromotionRepository {
	return &promotionRepository{
		db:        db,
		cacheRepo: cacheRepo,
	}
}

func (r *promotionRepository) CreatePromotion(ctx context.Context, promotion *model.Promotion) error {
	return r.db.WithContext(ctx).Create(promotion).Error
}

func (r *promotionRepository) GetPromotionByID(ctx context.Context, id string) (*model.Promotion, error) {
	var promotion model.Promotion
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&promotion).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &promotion, nil
}

func (r *promotionRepository) UpdatePromotion(ctx context.Context, promotion *model.Promotion) error {
	return r.db.WithContext(ctx).
		Model(&model.Promotion{}).
		Where("id = ?", promotion.ID).
		Updates(promotion).Error
}

func (r *promotionRepository) DeletePromotion(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Promotion{}).Error
}

func (r *promotionRepository) ListPromotions(ctx context.Context, filter *PromotionListFilter) ([]*model.Promotion, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Promotion{})

	if filter.Type > 0 {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Status > 0 {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Keyword != "" {
		safeKeyword := strings.ReplaceAll(filter.Keyword, `\`, `\\`)
		safeKeyword = strings.ReplaceAll(safeKeyword, "%", "\\%")
		safeKeyword = strings.ReplaceAll(safeKeyword, "_", "\\_")
		query = query.Where("name LIKE ?", "%"+safeKeyword+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var promotions []*model.Promotion
	err := query.
		Order("sort_order DESC, created_at DESC").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Find(&promotions).Error

	return promotions, total, err
}

func (r *promotionRepository) GetAvailablePromotions(ctx context.Context, productIDs []string, categoryID string, totalAmount float64) ([]*model.Promotion, error) {
	now := r.db.NowFunc()
	query := r.db.WithContext(ctx).Model(&model.Promotion{}).
		Where("status = ?", model.PromotionStatusActive).
		Where("start_time <= ?", now).
		Where("end_time >= ?", now)

	// 如果有商品ID，查询适用这些商品的促销
	if len(productIDs) > 0 {
		// 这里简化处理，实际应该解析 product_ids JSON 字段进行匹配
		// 或者使用 MySQL 的 JSON_CONTAINS 函数
		query = query.Where("product_ids = '' OR product_ids IS NULL") // 简化：只查询全平台促销
	}

	// 如果有类目ID，查询适用该类目的促销
	if categoryID != "" {
		query = query.Where("category_ids = '' OR category_ids IS NULL") // 简化：只查询全平台促销
	}

	var promotions []*model.Promotion
	err := query.Order("sort_order DESC").Find(&promotions).Error
	return promotions, err
}

func (r *promotionRepository) RecordPromotionUsage(ctx context.Context, log *model.PromotionUsageLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *promotionRepository) GetUserPromotionUsageCount(ctx context.Context, promotionID, userID string) (int32, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.PromotionUsageLog{}).
		Where("promotion_id = ? AND user_id = ?", promotionID, userID).
		Count(&count).Error
	return int32(count), err
}
