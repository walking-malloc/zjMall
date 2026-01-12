package model

import (
	"zjMall/pkg"

	"gorm.io/gorm"
)

// ProductTag 商品标签关联模型
// 对应数据库表：product_tags
type ProductTag struct {
	pkg.BaseModel
	ProductID string         `gorm:"type:varchar(26);comment:商品ID" json:"product_id"`
	TagID     string         `gorm:"type:varchar(26);comment:标签ID" json:"tag_id"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// TableName 指定表名
func (ProductTag) TableName() string {
	return "product_tags"
}
