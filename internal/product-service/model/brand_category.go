package model

import (
	"zjMall/pkg"

	"gorm.io/gorm"
)

// BrandCategory 品牌类目关联模型
// 对应数据库表：brand_categories
type BrandCategory struct {
	pkg.BaseModel
	BrandID    string         `gorm:"type:varchar(26);not null;index:idx_brand_id;comment:品牌ID" json:"brand_id"`
	CategoryID string         `gorm:"type:varchar(26);not null;index:idx_category_id;comment:类目ID" json:"category_id"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// TableName 指定表名
func (BrandCategory) TableName() string {
	return "brand_categories"
}
