package model

import "zjMall/pkg"

// OrderItem 订单明细表
type OrderItem struct {
	pkg.BaseModel

	OrderNo string `gorm:"type:varchar(32);index;not null;comment:订单号" json:"order_no"`
	UserID  string `gorm:"type:varchar(26);index;not null;comment:用户ID" json:"user_id"`

	ProductID    string  `gorm:"type:varchar(26);not null;comment:商品ID" json:"product_id"`
	SKUID        string  `gorm:"column:sku_id;type:varchar(26);not null;comment:SKU ID" json:"sku_id"`
	ProductTitle string  `gorm:"type:varchar(200);not null;comment:商品标题快照" json:"product_title"`
	ProductImage string  `gorm:"type:varchar(255);comment:商品图片快照" json:"product_image"`
	SKUName      string  `gorm:"type:varchar(100);comment:SKU 名称快照" json:"sku_name"`
	Price        float64 `gorm:"type:decimal(10,2);not null;comment:商品单价快照" json:"price"`
	Quantity     int32   `gorm:"type:int;not null;default:1;comment:购买数量" json:"quantity"`
	Subtotal     float64 `gorm:"type:decimal(10,2);not null;default:0;comment:小计金额" json:"subtotal"`
}

func (OrderItem) TableName() string {
	return "order_items"
}
