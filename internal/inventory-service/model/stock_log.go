package model

import (
	"time"

	"zjMall/pkg"

	"gorm.io/gorm"
)

// StockLog 库存变动明细
// 对应表：inventory_logs
type StockLog struct {
	ID           string    `gorm:"type:varchar(26);primaryKey;comment:日志ID"`
	SKUID        string    `gorm:"column:sku_id;type:varchar(26);index;not null;comment:SKU ID" json:"sku_id"`
	ChangeAmount int64     `gorm:"type:int;not null;comment:库存变动数量：正数增加，负数减少" json:"change_amount"`
	Reason       string    `gorm:"type:varchar(50);not null;comment:变动原因" json:"reason"`
	RefID        string    `gorm:"type:varchar(64);comment:关联单号（订单号/操作单号等）" json:"ref_id"`
	CreatedAt    time.Time `gorm:"comment:创建时间" json:"created_at"`
}

func (StockLog) TableName() string {
	return "inventory_logs"
}

// BeforeCreate GORM 钩子，在插入前自动生成主键 ID（适配 GORM v2 的签名：BeforeCreate(*gorm.DB) error）
func (s *StockLog) BeforeCreate(tx *gorm.DB) (err error) {
	// 只有在 ID 为空时才生成，避免调用方显式指定 ID 时被覆盖
	if s.ID == "" {
		s.ID = pkg.GenerateULID()
	}
	return nil
}
