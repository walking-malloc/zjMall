package model

import "zjMall/pkg"

// Address 收货地址模型
// 对应数据库表：addresses
type Address struct {
	pkg.BaseModel

	// 关联信息
	UserID string `json:"user_id" gorm:"type:varchar(26);index"`

	// 收货人信息
	ReceiverName  string `json:"receiver_name"`
	ReceiverPhone string `json:"receiver_phone"`

	// 地址信息
	Province   string `json:"province"`
	City       string `json:"city"`
	District   string `json:"district"`
	Detail     string `json:"detail"`
	PostalCode string `json:"postal_code,omitempty"` // 邮政编码（可选）

	// 状态信息
	IsDefault bool `json:"is_default"` // 是否默认：0-否，1-是
}

// TableName 指定表名（GORM 约定：默认使用结构体名的复数形式）
func (Address) TableName() string {
	return "addresses"
}
