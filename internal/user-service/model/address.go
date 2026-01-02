package model

import "zjMall/pkg"

// Address 收货地址模型
// 对应数据库表：addresses
type Address struct {
	pkg.BaseModel

	// 关联信息
	UserID string `gorm:"type:varchar(26);not null;index:idx_user_default;comment:用户ID" json:"user_id"`

	// 收货人信息
	ReceiverName  string `gorm:"type:varchar(50);not null;comment:收货人姓名" json:"receiver_name"`
	ReceiverPhone string `gorm:"type:varchar(11);not null;comment:收货人手机号" json:"receiver_phone"`

	// 地址信息
	Province   string `gorm:"type:varchar(50);not null;comment:省份" json:"province"`
	City       string `gorm:"type:varchar(50);not null;comment:城市" json:"city"`
	District   string `gorm:"type:varchar(50);not null;comment:区县" json:"district"`
	Detail     string `gorm:"type:varchar(200);not null;comment:详细地址" json:"detail"`
	PostalCode string `gorm:"type:varchar(6);comment:邮政编码" json:"postal_code,omitempty"` // 邮政编码（可选）

	// 状态信息
	IsDefault bool `gorm:"type:tinyint(1);default:0;index:idx_user_default;comment:是否默认：0-否，1-是" json:"is_default"`
}

// TableName 指定表名（GORM 约定：默认使用结构体名的复数形式）
func (Address) TableName() string {
	return "addresses"
}
