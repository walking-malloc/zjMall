package model

import (
	"time"
	"zjMall/pkg"
)

// User 用户模型
// 对应数据库表：users
type User struct {
	pkg.BaseModel

	// 账号信息
	Phone    string `json:"phone"`
	Password string `json:"-"` // json:"-" 表示不序列化到 JSON

	// 基本信息
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Email    string `json:"email"`

	// 个人信息
	Gender   int8       `json:"gender"`
	Birthday *time.Time `json:"birthday,omitempty"` // 使用指针，nil 表示未设置

	// 状态信息
	Status int8 `json:"status"`

	// 登录信息
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

// TableName 指定表名（GORM 约定：默认使用结构体名的复数形式）
func (User) TableName() string {
	return "users"
}
