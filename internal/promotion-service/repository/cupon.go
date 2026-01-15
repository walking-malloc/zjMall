ackage repository

import (
	"context"
	"errors"
	"strings"
	"time"
	"zjMall/internal/common/cache"
	"zjMall/internal/promotion-service/model"

	"gorm.io/gorm"
)

type CouponListFilter struct {
	Page     int32
	PageSize int32
	UserID   string
	Status   int8
	Offset   int
	Limit    int
}

type CouponRepository interface {
	// 优惠券模板
	CreateCouponTemplate(ctx context.Context, template *model.CouponTemplate) error
	GetCouponTemplateByID(ctx context.Context, id string) (*model.CouponTemplate, error)
	UpdateCouponTemplate(ctx context.Context, template *model.CouponTemplate) error
	ListCouponTemplates(ctx context.Context, page, pageSize int32) ([]*model.CouponTemplate, int64, error)
	
	// 优惠券实例
	CreateCoupon(ctx context.Context, coupon *model.Coupon) error
	GetCouponByID(ctx context.Context, id string) (*model.Coupon, error)
	GetCouponByUserAndTemplate(ctx context.Context, userID, templateID string) ([]*model.Coupon, error)
	ListUserCoupons(ctx context.Context, filter *CouponListFilter) ([]*model.Coupon, int64, error)
	UpdateCoupon(ctx context.Context, coupon *model.Coupon) error
	
	// 统计
	GetUserCouponCountByTemplate(ctx context.Context, userID, templateID string) (int32, error)
	IncrementTemplateClaimedCount(ctx context.Context, templateID string) error
}

type couponRepository struct {
	db        *gorm.DB
	cacheRepo cache.CacheRepository
}

func NewCouponRepository(db *gorm.DB, cacheRepo cache.CacheRepository) CouponRepository {
	return &couponRepository{
		db:        db,
		cacheRepo: cacheRepo,
	}
}

func (r *couponRepository) CreateCouponTemplate(ctx context.Context, template *model.CouponTemplate) error {
	return r.db.WithContext(ctx).Create(template).Error
}

func (r *couponRepository) GetCouponTemplateByID(ctx context.Context, id string) (*model.CouponTemplate, error) {
	var template model.CouponTemplate
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&template).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

func (r *couponRepository) UpdateCouponTemplate(ctx context.Context, template *model.CouponTemplate) error {
	return r.db.WithContext(ctx).
		Model(&model.CouponTemplate{}).
		Where("id = ?", template.ID).
		Updates(template).Error
}

func (r *couponRepository) ListCouponTemplates(ctx context.Context, page, pageSize int32) ([]*model.CouponTemplate, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.CouponTemplate{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var templates []*model.CouponTemplate
	offset := int((page - 1) * pageSize)
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(int(pageSize)).
		Find(&templates).Error

	return templates, total, err
}

func (r *couponRepository) CreateCoupon(ctx context.Context, coupon *model.Coupon) error {
	return r.db.WithContext(ctx).Create(coupon).Error
}

func (r *couponRepository) GetCouponByID(ctx context.Context, id string) (*model.Coupon, error) {
	var coupon model.Coupon
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&coupon).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &coupon, nil
}

func (r *couponRepository) GetCouponByUserAndTemplate(ctx context.Context, userID, templateID string) ([]*model.Coupon, error) {
	var coupons []*model.Coupon
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND template_id = ?", userID, templateID).
		Find(&coupons).Error
	return coupons, err
}

func (r *couponRepository) ListUserCoupons(ctx context.Context, filter *CouponListFilter) ([]*model.Coupon, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Coupon{}).
		Where("user_id = ?", filter.UserID)

	if filter.Status > 0 {
		query = query.Where("status = ?", filter.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var coupons []*model.Coupon
	err := query.
		Order("status ASC, valid_end_time ASC").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Find(&coupons).Error

	return coupons, total, err
}

func (r *couponRepository) UpdateCoupon(ctx context.Context, coupon *model.Coupon) error {
	return r.db.WithContext(ctx).
		Model(&model.Coupon{}).
		Where("id = ?", coupon.ID).
		Updates(coupon).Error
}

func (r *couponRepository) GetUserCouponCountByTemplate(ctx context.Context, userID, templateID string) (int32, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Coupon{}).
		Where("user_id = ? AND template_id = ?", userID, templateID).
		Count(&count).Error
	return int32(count), err
}

func (r *couponRepository) IncrementTemplateClaimedCount(ctx context.Context, templateID string) error {
	return r.db.WithContext(ctx).
		Model(&model.CouponTemplate{}).
		Where("id = ?", templateID).
		UpdateColumn("claimed_count", gorm.Expr("claimed_count + ?", 1)).Error
}