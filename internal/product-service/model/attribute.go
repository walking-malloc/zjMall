package model

import (
	"zjMall/pkg"

	"gorm.io/gorm"
)

// Attribute 属性模型（属性模板）
// 对应数据库表：attributes
type Attribute struct {
	pkg.BaseModel

	// 关联信息
	CategoryID string `gorm:"type:varchar(26);not null;comment:所属类目ID" json:"category_id"`

	// 基本信息
	Name      string `gorm:"type:varchar(100);not null;comment:属性名称（如：颜色、尺寸、存储容量）" json:"name"`
	Type      int8   `gorm:"type:tinyint;not null;default:1;comment:属性类型：1-销售属性（用于生成SKU），2-非销售属性（仅展示）" json:"type"`
	InputType int8   `gorm:"type:tinyint;not null;default:1;comment:录入方式：1-单选，2-多选，3-文本，4-数值" json:"input_type"`

	// 业务规则
	IsRequired int8 `gorm:"type:tinyint;default:0;comment:是否必填：0-否，1-是" json:"is_required"`

	// 排序
	SortOrder int32          `gorm:"type:int;default:0;comment:排序权重" json:"sort_order"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// TableName 指定表名
func (Attribute) TableName() string {
	return "attributes"
}
