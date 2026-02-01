package model

import (
	"time"
	"zjMall/pkg"
)

// Refund 退款单表
// 用于记录退款信息，与支付单关联
type Refund struct {
	pkg.BaseModel

	RefundNo  string `gorm:"type:varchar(32);uniqueIndex;not null;comment:退款单号" json:"refund_no"`
	PaymentNo string `gorm:"type:varchar(32);index;not null;comment:原支付单号" json:"payment_no"`
	OrderNo   string `gorm:"type:varchar(32);index;not null;comment:订单号" json:"order_no"`
	UserID    string `gorm:"type:varchar(26);index;not null;comment:用户ID" json:"user_id"`

	RefundAmount float64 `gorm:"type:decimal(10,2);not null;default:0;comment:退款金额" json:"refund_amount"`
	RefundReason string  `gorm:"type:varchar(255);comment:退款原因" json:"refund_reason"`
	RefundType   int8    `gorm:"type:tinyint;not null;default:1;comment:退款类型：1-全额退款，2-部分退款" json:"refund_type"`

	PayChannel string `gorm:"type:varchar(20);not null;comment:原支付渠道" json:"pay_channel"`
	Status     int8   `gorm:"type:tinyint;not null;default:1;comment:退款状态：1-退款中，2-退款成功，3-退款失败，4-已取消" json:"status"`

	TradeNo       string `gorm:"type:varchar(64);comment:原支付交易号" json:"trade_no"`
	RefundTradeNo string `gorm:"type:varchar(64);index;comment:退款交易号（第三方返回）" json:"refund_trade_no"`

	RequestData  string `gorm:"type:text;comment:退款请求数据（JSON格式）" json:"request_data"`
	ResponseData string `gorm:"type:text;comment:退款响应数据（JSON格式）" json:"response_data"`
	ErrorMessage string `gorm:"type:varchar(500);comment:错误信息" json:"error_message"`

	RefundedAt *time.Time `gorm:"type:timestamp;null;default:null;comment:退款成功时间" json:"refunded_at"`
	Version    int        `gorm:"type:int;not null;default:0;comment:版本号（乐观锁）" json:"version"`
}

func (Refund) TableName() string {
	return "refunds"
}

// 退款类型常量
const (
	RefundTypeFull    = int8(1) // 全额退款
	RefundTypePartial = int8(2) // 部分退款
)

// 退款状态常量
const (
	RefundStatusProcessing = int8(1) // 退款中
	RefundStatusSuccess    = int8(2) // 退款成功
	RefundStatusFailed     = int8(3) // 退款失败
	RefundStatusCancelled  = int8(4) // 已取消
)
