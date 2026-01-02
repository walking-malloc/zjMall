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
	Phone    string `gorm:"type:varchar(11);uniqueIndex;not null;comment:手机号" json:"phone"`
	Password string `gorm:"type:varchar(255);not null;comment:密码（加密）" json:"-"` // json:"-" 表示不序列化到 JSON

	// 基本信息
	Nickname string `gorm:"type:varchar(50);comment:昵称" json:"nickname"`
	Avatar   string `gorm:"type:varchar(255);comment:头像URL" json:"avatar"`
	Email    string `gorm:"type:varchar(100);comment:邮箱" json:"email"`

	// 个人信息
	Gender   int8       `gorm:"type:tinyint(1);default:0;comment:性别：0-未设置，1-男，2-女" json:"gender"`
	Birthday *time.Time `gorm:"type:date;comment:生日" json:"birthday,omitempty"` // 使用指针，nil 表示未设置

	// 状态信息
	Status int8 `gorm:"type:tinyint(1);default:1;comment:状态：1-正常，2-已锁定，3-已注销" json:"status"`

	// 登录信息
	LastLoginAt *time.Time `gorm:"type:timestamp;comment:最后登录时间" json:"last_login_at,omitempty"`
}

// TableName 指定表名（GORM 约定：默认使用结构体名的复数形式）
func (User) TableName() string {
	return "users"
}
