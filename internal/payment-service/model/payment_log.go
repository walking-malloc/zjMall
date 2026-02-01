package model

import (
	"zjMall/pkg"
)

// PaymentLog 支付日志表
// 用于记录支付操作的详细日志，便于审计与排查问题
type PaymentLog struct {
	pkg.BaseModel

	PaymentNo string `gorm:"type:varchar(32);index;not null;comment:支付单号" json:"payment_no"`
	OrderNo   string `gorm:"type:varchar(32);index;not null;comment:订单号" json:"order_no"`
	UserID    string `gorm:"type:varchar(26);index;comment:用户ID" json:"user_id"`

	Action     string `gorm:"type:varchar(50);not null;comment:操作类型" json:"action"`
	FromStatus *int8  `gorm:"type:tinyint;comment:变更前状态" json:"from_status"`
	ToStatus   *int8  `gorm:"type:tinyint;comment:变更后状态" json:"to_status"`

	Channel string  `gorm:"type:varchar(20);comment:支付渠道" json:"channel"`
	Amount  float64 `gorm:"type:decimal(10,2);comment:支付金额" json:"amount"`
	TradeNo string  `gorm:"type:varchar(64);comment:第三方交易号" json:"trade_no"`

	RequestData  string `gorm:"type:text;comment:请求数据（JSON格式）" json:"request_data"`
	ResponseData string `gorm:"type:text;comment:响应数据（JSON格式）" json:"response_data"`
	ErrorMessage string `gorm:"type:varchar(500);comment:错误信息" json:"error_message"`

	IPAddress string `gorm:"type:varchar(50);comment:IP地址" json:"ip_address"`
	UserAgent string `gorm:"type:varchar(255);comment:用户代理" json:"user_agent"`
}

func (PaymentLog) TableName() string {
	return "payment_logs"
}

// 操作类型常量
const (
	PaymentLogActionCreate       = "create"        // 创建支付单
	PaymentLogActionCallback     = "callback"      // 支付回调
	PaymentLogActionStatusChange = "status_change" // 状态变更
	PaymentLogActionClose        = "close"         // 关闭支付单
	PaymentLogActionQuery        = "query"         // 查询支付状态
)
