package model

import (
	"zjMall/pkg"
)

// Role 角色模型
type Role struct {
	pkg.BaseModel
	Code        string `gorm:"type:varchar(50);uniqueIndex;not null;comment:角色代码" json:"code"`
	Name        string `gorm:"type:varchar(100);not null;comment:角色名称" json:"name"`
	Description string `gorm:"type:varchar(255);comment:角色描述" json:"description"`
	Status      int8   `gorm:"type:tinyint(1);default:1;comment:状态：1-启用，2-停用" json:"status"`
}

func (Role) TableName() string {
	return "roles"
}

// UserRole 用户角色关联模型
type UserRole struct {
	pkg.BaseModel
	UserID string `gorm:"type:varchar(26);not null;index;comment:用户ID" json:"user_id"`
	RoleID string `gorm:"type:varchar(26);not null;index;comment:角色ID" json:"role_id"`
}

func (UserRole) TableName() string {
	return "user_roles"
}

// 之前这里有 Permission / RolePermission 结构体，目前已改用 Casbin 配置做权限控制，因此删除
