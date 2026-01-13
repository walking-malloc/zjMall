package model

import "zjMall/pkg"

// SkuAttribute SKU 属性关联模型
// 对应数据库表：sku_attributes
type SkuAttribute struct {
	pkg.BaseModel

	SkuID            string `gorm:"type:varchar(26);not null;index:idx_sku_id;comment:SKU ID" json:"sku_id"`
	AttributeValueID string `gorm:"type:varchar(26);not null;index:idx_attribute_value_id;comment:属性值ID" json:"attribute_value_id"`
}

// TableName 指定表名
func (SkuAttribute) TableName() string {
	return "sku_attributes"
}
