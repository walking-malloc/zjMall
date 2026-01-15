package model

import (
	"time"
	"zjMall/pkg"

	"gorm.io/gorm"
)

// PromotionType 促销类型
const (
	PromotionTypeFullReduction   int8 = 1 // 满减
	PromotionTypeFullDiscount    int8 = 2 // 满折
	PromotionTypeDirectReduction int8 = 3 // 直降
	PromotionTypeTimeLimited     int8 = 4 // 限时折扣
)

// PromotionStatus 促销状态
const (
	PromotionStatusDraft   int8 = 1 // 草稿
	PromotionStatusActive  int8 = 2 // 进行中
	PromotionStatusPaused  int8 = 3 // 已暂停
	PromotionStatusEnded   int8 = 4 // 已结束
	PromotionStatusDeleted int8 = 5 // 已删除
)

// Promotion 促销活动模型
type Promotion struct {
	pkg.BaseModel

	Name           string    `gorm:"type:varchar(100);not null;comment:促销名称" json:"name"`
	Type           int8      `gorm:"type:tinyint;not null;comment:促销类型：1-满减，2-满折，3-直降，4-限时折扣" json:"type"`
	Description    string    `gorm:"type:text;comment:促销描述" json:"description,omitempty"`
	ProductIDs     string    `gorm:"type:text;comment:适用商品ID列表（JSON数组）" json:"product_ids,omitempty"`
	CategoryIDs    string    `gorm:"type:text;comment:适用类目ID列表（JSON数组）" json:"category_ids,omitempty"`
	ConditionValue string    `gorm:"type:varchar(50);comment:条件值（如：满200）" json:"condition_value,omitempty"`
	DiscountValue  string    `gorm:"type:varchar(50);comment:优惠值（如：减30 或 打8折）" json:"discount_value,omitempty"`
	StartTime      time.Time `gorm:"type:timestamp;not null;comment:开始时间" json:"start_time"`
	EndTime        time.Time `gorm:"type:timestamp;not null;comment:结束时间" json:"end_time"`
	MaxUseTimes    int32     `gorm:"type:int;default:0;comment:每人限用次数（0表示不限制）" json:"max_use_times"`
	TotalQuota     int32     `gorm:"type:int;default:0;comment:总配额（0表示不限制）" json:"total_quota"`
	UsedQuota      int32     `gorm:"type:int;default:0;comment:已使用配额" json:"used_quota"`
	SortOrder      int32     `gorm:"type:int;default:0;comment:排序权重" json:"sort_order"`
	Status         int8      `gorm:"type:tinyint;default:1;comment:状态：1-草稿，2-进行中，3-已暂停，4-已结束，5-已删除" json:"status"`

	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (Promotion) TableName() string {
	return "promotions"
}

// PromotionUsageLog 促销使用记录
type PromotionUsageLog struct {
	pkg.BaseModel

	PromotionID    string  `gorm:"type:varchar(26);not null;index:idx_promotion_user;comment:促销活动ID" json:"promotion_id"`
	UserID         string  `gorm:"type:varchar(26);not null;index:idx_promotion_user;comment:用户ID" json:"user_id"`
	OrderID        string  `gorm:"type:varchar(26);index;comment:订单ID" json:"order_id,omitempty"`
	DiscountAmount float64 `gorm:"type:decimal(10,2);comment:优惠金额" json:"discount_amount"`
}

func (PromotionUsageLog) TableName() string {
	return "promotion_usage_logs"
}
