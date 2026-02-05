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

// Permission 权限模型
type Permission struct {
	pkg.BaseModel
	Code        string `gorm:"type:varchar(100);uniqueIndex;not null;comment:权限代码" json:"code"`
	Name        string `gorm:"type:varchar(100);not null;comment:权限名称" json:"name"`
	Resource    string `gorm:"type:varchar(50);not null;comment:资源类型" json:"resource"`
	Action      string `gorm:"type:varchar(50);not null;comment:操作类型" json:"action"`
	Description string `gorm:"type:varchar(255);comment:权限描述" json:"description"`
	Status      int8   `gorm:"type:tinyint(1);default:1;comment:状态：1-启用，2-停用" json:"status"`
}

func (Permission) TableName() string {
	return "permissions"
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

// RolePermission 角色权限关联模型
type RolePermission struct {
	pkg.BaseModel
	RoleID       string `gorm:"type:varchar(26);not null;index;comment:角色ID" json:"role_id"`
	PermissionID string `gorm:"type:varchar(26);not null;index;comment:权限ID" json:"permission_id"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}
