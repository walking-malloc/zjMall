package model

import (
	"time"
	"zjMall/pkg"
)

// CouponType 优惠券类型
const (
	CouponTypeFixed        int8 = 1 // 固定金额券
	CouponTypePercent      int8 = 2 // 折扣券
	CouponTypeFreeShipping int8 = 3 // 免运费券
)

// CouponStatus 优惠券状态
const (
	CouponStatusUnused  int8 = 1 // 未使用
	CouponStatusUsed    int8 = 2 // 已使用
	CouponStatusExpired int8 = 3 // 已过期
)

// CouponTemplate 优惠券模板
type CouponTemplate struct {
	pkg.BaseModel

	Name           string    `gorm:"type:varchar(100);not null;comment:优惠券名称" json:"name"`
	Type           int8      `gorm:"type:tinyint;not null;comment:优惠券类型：1-固定金额，2-折扣，3-免运费" json:"type"`
	Description    string    `gorm:"type:text;comment:优惠券描述" json:"description,omitempty"`
	DiscountValue  string    `gorm:"type:varchar(50);not null;comment:优惠值" json:"discount_value"`
	ConditionValue string    `gorm:"type:varchar(50);comment:使用条件（如：满100可用）" json:"condition_value,omitempty"`
	TotalCount     int32     `gorm:"type:int;default:0;comment:发放总数（0表示不限制）" json:"total_count"`
	ClaimedCount   int32     `gorm:"type:int;default:0;comment:已领取数量" json:"claimed_count"`
	PerUserLimit   int32     `gorm:"type:int;default:1;comment:每人限领数量" json:"per_user_limit"`
	ValidStartTime time.Time `gorm:"type:timestamp;not null;comment:有效期开始时间" json:"valid_start_time"`
	ValidEndTime   time.Time `gorm:"type:timestamp;not null;comment:有效期结束时间" json:"valid_end_time"`
	ValidDays      int32     `gorm:"type:int;default:0;comment:领取后有效天数（0表示使用模板有效期）" json:"valid_days"`
	Status         int8      `gorm:"type:tinyint;default:1;comment:状态：1-启用，2-停用" json:"status"`
}

func (CouponTemplate) TableName() string {
	return "coupon_templates"
}

// Coupon 优惠券实例（用户领取的优惠券）
type Coupon struct {
	pkg.BaseModel

	TemplateID     string     `gorm:"type:varchar(26);not null;index;comment:优惠券模板ID" json:"template_id"`
	UserID         string     `gorm:"type:varchar(26);not null;index:idx_user_status;comment:用户ID" json:"user_id"`
	Name           string     `gorm:"type:varchar(100);not null;comment:优惠券名称" json:"name"`
	Type           int8       `gorm:"type:tinyint;not null;comment:优惠券类型" json:"type"`
	Description    string     `gorm:"type:text;comment:优惠券描述" json:"description,omitempty"`
	DiscountValue  string     `gorm:"type:varchar(50);not null;comment:优惠值" json:"discount_value"`
	ConditionValue string     `gorm:"type:varchar(50);comment:使用条件" json:"condition_value,omitempty"`
	Status         int8       `gorm:"type:tinyint;default:1;comment:状态：1-未使用，2-已使用，3-已过期" json:"status"`
	ValidStartTime time.Time  `gorm:"type:timestamp;not null;comment:有效期开始时间" json:"valid_start_time"`
	ValidEndTime   time.Time  `gorm:"type:timestamp;not null;comment:有效期结束时间" json:"valid_end_time"`
	UsedAt         *time.Time `gorm:"type:timestamp;comment:使用时间" json:"used_at,omitempty"`
	OrderID        string     `gorm:"type:varchar(26);comment:使用的订单ID" json:"order_id,omitempty"`
}

func (Coupon) TableName() string {
	return "coupons"
}
