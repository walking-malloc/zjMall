package model

import (
	"time"
)

// Stock 库存主表模型
// 建议对应表名：inventory_stocks
type Stock struct {
	ID             string    `gorm:"type:varchar(26);primaryKey;comment:主键ID"`
	SKUID          string    `gorm:"type:varchar(26);uniqueIndex;not null;comment:SKU ID" json:"sku_id"`
	AvailableStock int64     `gorm:"type:int;not null;default:0;comment:可用库存" json:"available_stock"`
	Version        int64     `gorm:"type:bigint;not null;default:0;comment:乐观锁版本号" json:"version"`
	CreatedAt      time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt      time.Time `gorm:"comment:更新时间" json:"updated_at"`
}

// TableName 指定库存表名
func (Stock) TableName() string {
	return "inventory_stocks"
}
