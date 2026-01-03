package model

import (
	"time"
	"zjMall/pkg"
)

// Category 类目模型
// 对应数据库表：categories
type Category struct {
	pkg.BaseModel

	// 树形结构
	ParentID string `gorm:"type:varchar(26);index:idx_parent_visible_status;comment:父类目ID，顶级类目为NULL" json:"parent_id,omitempty"`

	// 基本信息
	Name      string `gorm:"type:varchar(100);not null;comment:类目名称" json:"name"`
	Level     int8   `gorm:"type:tinyint;not null;default:1;index:idx_level_status;comment:类目层级：1-一级，2-二级，3-三级" json:"level"`
	IsLeaf    bool   `gorm:"type:tinyint(1);default:0;comment:是否为叶子节点：0-否，1-是" json:"is_leaf"`
	IsVisible bool   `gorm:"type:tinyint(1);default:1;index:idx_parent_visible_status;comment:是否在前台展示：0-否，1-是" json:"is_visible"`

	// 展示相关
	SortOrder int32  `gorm:"type:int;default:0;comment:排序权重，数字越大越靠前" json:"sort_order"`
	Icon      string `gorm:"type:varchar(255);comment:类目图标URL" json:"icon,omitempty"`

	// 状态
	Status int8 `gorm:"type:tinyint(1);default:1;index:idx_level_status,idx_parent_visible_status;comment:状态：1-启用，2-停用" json:"status"`

	// 软删除
	DeletedAt time.Time `gorm:"type:timestamp;index;comment:软删除时间" json:"deleted_at,omitempty"`
}

// TableName 指定表名
func (Category) TableName() string {
	return "categories"
}
