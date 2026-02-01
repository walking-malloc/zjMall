package model

import (
	"zjMall/pkg"
)

// PaymentChannel 支付渠道配置表
// 用于管理支付渠道的配置信息（商户号、密钥等）
// 支持多商户、多应用场景
type PaymentChannel struct {
	pkg.BaseModel

	ChannelCode string `gorm:"type:varchar(20);index;not null;comment:渠道代码" json:"channel_code"`
	ChannelName string `gorm:"type:varchar(50);not null;comment:渠道名称" json:"channel_name"`

	AppID      string `gorm:"type:varchar(64);comment:应用ID（微信AppID/支付宝AppID）" json:"app_id"`
	MerchantID string `gorm:"type:varchar(64);comment:商户号" json:"merchant_id"`
	MchID      string `gorm:"type:varchar(64);comment:微信商户号（微信专用）" json:"mch_id"`

	APIKey     string `gorm:"type:varchar(255);comment:API密钥（加密存储）" json:"api_key"`
	PublicKey  string `gorm:"type:text;comment:公钥（支付宝专用）" json:"public_key"`
	PrivateKey string `gorm:"type:text;comment:私钥（支付宝专用，加密存储）" json:"private_key"`

	NotifyURL string `gorm:"type:varchar(255);comment:回调地址" json:"notify_url"`
	ReturnURL string `gorm:"type:varchar(255);comment:返回地址" json:"return_url"`

	IsEnabled   bool   `gorm:"type:tinyint(1);not null;default:1;comment:是否启用：0-禁用，1-启用" json:"is_enabled"`
	IsDefault   bool   `gorm:"type:tinyint(1);not null;default:0;comment:是否默认渠道：0-否，1-是" json:"is_default"`
	Environment string `gorm:"type:varchar(20);not null;default:'sandbox';index;comment:环境：sandbox-沙箱，production-生产" json:"environment"`

	Remark string `gorm:"type:varchar(255);comment:备注" json:"remark"`
}

func (PaymentChannel) TableName() string {
	return "payment_channels"
}

// 环境常量
const (
	EnvironmentSandbox    = "sandbox"    // 沙箱环境
	EnvironmentProduction = "production" // 生产环境
)
