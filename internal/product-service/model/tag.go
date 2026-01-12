package model

import (
	"zjMall/pkg"

	"gorm.io/gorm"
)

// Tag 标签模型
// 对应数据库表：tags
type Tag struct {
	pkg.BaseModel

	Name      string `gorm:"type:varchar(50);not null;comment:标签名称" json:"name"`
	Type      int8   `gorm:"type:tinyint;not null;default:2;comment:标签类型：1-系统标签，2-运营标签" json:"type"`
	Color     string `gorm:"type:varchar(20);comment:标签颜色（用于前端展示）" json:"color,omitempty"`
	SortOrder int32  `gorm:"type:int;not null;default:0;comment:排序权重" json:"sort_order"`
	Status    int8   `gorm:"type:tinyint;not null;default:1;comment:状态：1-启用，2-停用" json:"status"`

	// 软删除
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// TableName 指定表名
func (Tag) TableName() string {
	return "tags"
}
