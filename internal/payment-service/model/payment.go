package model

import (
	"time"
	"zjMall/pkg"
)

// Payment 支付单表
type Payment struct {
	pkg.BaseModel

	PaymentNo string  `gorm:"type:varchar(32);uniqueIndex;not null;comment:支付单号" json:"payment_no"`
	OrderNo   string  `gorm:"type:varchar(32);index;not null;comment:订单号" json:"order_no"`
	UserID    string  `gorm:"type:varchar(26);index;not null;comment:用户ID" json:"user_id"`
	Amount    float64 `gorm:"type:decimal(10,2);not null;default:0;comment:支付金额" json:"amount"`

	PayChannel string `gorm:"type:varchar(20);not null;comment:支付渠道(wechat/alipay)" json:"pay_channel"`
	Status     int8   `gorm:"type:tinyint;not null;default:1;comment:支付状态(1待支付/2支付中/3成功/4失败/5已关闭/6已退款)" json:"status"`

	TradeNo   string `gorm:"type:varchar(64);index;comment:第三方交易号" json:"trade_no"`
	NotifyURL string `gorm:"type:varchar(255);comment:回调地址" json:"notify_url"`
	ReturnURL string `gorm:"type:varchar(255);comment:返回地址" json:"return_url"`

	PaidAt    *time.Time `gorm:"type:timestamp;null;default:null;comment:支付时间" json:"paid_at"`
	ExpiredAt *time.Time `gorm:"type:timestamp;null;default:null;comment:过期时间" json:"expired_at"`

	Version int `gorm:"type:int;not null;default:0;comment:版本号(乐观锁)" json:"version"`
}

func (Payment) TableName() string {
	return "payments"
}

// 支付状态常量
const (
	PaymentStatusPending    = int8(1) // 待支付
	PaymentStatusProcessing = int8(2) // 支付中
	PaymentStatusSuccess    = int8(3) // 支付成功
	PaymentStatusFailed     = int8(4) // 支付失败
	PaymentStatusClosed     = int8(5) // 已关闭
	PaymentStatusRefunded   = int8(6) // 已退款
)

// 支付渠道常量
const (
	PayChannelWeChat  = "wechat"
	PayChannelAlipay  = "alipay"
	PayChannelBalance = "balance"
)

// 支付单号前缀常量
const (
	PaymentNoPrefix = "10" // 支付单号前缀，用于区分订单号(01/02)和支付单号(10)
)
