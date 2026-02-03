package model

import (
	"time"
	"zjMall/pkg"
)

const (
	OrderTypeNormalPrefix  = "01"
	OrderTypeSeckillPrefix = "02"
	OrderTypeNormal        = "normal"
	OrderTypeSeckill       = "seckill"
)

// Order 订单主表
type Order struct {
	pkg.BaseModel

	OrderNo        string  `gorm:"type:varchar(32);uniqueIndex;not null;comment:订单号" json:"order_no"`
	UserID         string  `gorm:"type:varchar(26);index;not null;comment:用户ID" json:"user_id"`
	Status         int8    `gorm:"type:tinyint;not null;default:1;comment:订单状态" json:"status"`
	TotalAmount    float64 `gorm:"type:decimal(10,2);not null;default:0;comment:商品总金额" json:"total_amount"`
	DiscountAmount float64 `gorm:"type:decimal(10,2);not null;default:0;comment:优惠总金额" json:"discount_amount"`
	ShippingAmount float64 `gorm:"type:decimal(10,2);not null;default:0;comment:运费金额" json:"shipping_amount"`
	PayAmount      float64 `gorm:"type:decimal(10,2);not null;default:0;comment:应付金额" json:"pay_amount"`

	ReceiverName    string `gorm:"type:varchar(50);comment:收货人姓名" json:"receiver_name"`
	ReceiverPhone   string `gorm:"type:varchar(20);comment:收货人电话" json:"receiver_phone"`
	ReceiverAddress string `gorm:"type:varchar(255);comment:收货地址" json:"receiver_address"`

	BuyerRemark string `gorm:"type:varchar(255);comment:买家留言" json:"buyer_remark"`

	PayChannel    string `gorm:"type:varchar(20);comment:支付渠道" json:"pay_channel"`
	PayTradeNo    string `gorm:"type:varchar(64);comment:支付流水号" json:"pay_trade_no"`
	ItemsSnapshot string `gorm:"column:items_snapshot;type:json;comment:商品列表精简快照（JSON格式）" json:"items_snapshot"`

	CreatedAt   time.Time  `json:"created_at"`
	PaidAt      *time.Time `gorm:"type:timestamp;null;default:null;comment:支付时间" json:"paid_at"`
	ShippedAt   *time.Time `gorm:"type:timestamp;null;default:null;comment:发货时间" json:"shipped_at"`
	CompletedAt *time.Time `gorm:"type:timestamp;null;default:null;comment:完成时间" json:"completed_at"`
	Version     int        `gorm:"type:int;not null;default:0;comment:版本号" json:"version"`
}

func (Order) TableName() string {
	return "orders"
}
