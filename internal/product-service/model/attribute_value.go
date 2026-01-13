package model

import (
	"zjMall/pkg"

	"gorm.io/gorm"
)

// AttributeValue 属性值模型
// 对应数据库表：attribute_values
type AttributeValue struct {
	pkg.BaseModel

	AttributeID string         `gorm:"type:varchar(26);not null;index:idx_attribute_sort;comment:所属属性ID" json:"attribute_id"`
	Value       string         `gorm:"type:varchar(100);not null;comment:属性值名称" json:"value"`
	SortOrder   int32          `gorm:"type:int;not null;default:0;comment:排序权重" json:"sort_order"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// TableName 指定表名
func (AttributeValue) TableName() string {
	return "attribute_values"
}
