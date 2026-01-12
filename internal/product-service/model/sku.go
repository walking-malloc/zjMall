package model

import (
	"zjMall/pkg"

	"gorm.io/gorm"
)

// Sku SKU 模型（库存单元）
// 对应数据库表：skus
type Sku struct {
	pkg.BaseModel

	// 关联信息
	ProductID string `gorm:"type:varchar(26);not null;comment:所属商品ID（SPU）" json:"product_id"`

	// 基本信息
	SkuCode       string  `gorm:"type:varchar(50);uniqueIndex;comment:SKU编码（内部编码）" json:"sku_code,omitempty"`
	Barcode       string  `gorm:"type:varchar(50);comment:条形码" json:"barcode,omitempty"`
	Name          string  `gorm:"type:varchar(200);comment:SKU名称（如：黑色 128G）" json:"name,omitempty"`
	Price         float64 `gorm:"type:decimal(10,2);not null;comment:销售价格" json:"price"`
	OriginalPrice float64 `gorm:"type:decimal(10,2);comment:划线价/原价" json:"original_price,omitempty"`
	CostPrice     float64 `gorm:"type:decimal(10,2);comment:成本价" json:"cost_price,omitempty"`
	Weight        float64 `gorm:"type:decimal(10,2);comment:重量（kg）" json:"weight,omitempty"`
	Volume        float64 `gorm:"type:decimal(10,2);comment:体积（m³）" json:"volume,omitempty"`
	Image         string  `gorm:"type:varchar(255);comment:SKU图片" json:"image,omitempty"`

	// 状态
	Status int8 `gorm:"type:tinyint;not null;default:1;comment:状态：1-上架，2-下架，3-禁用" json:"status"`

	// 软删除
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// TableName 指定表名
func (Sku) TableName() string {
	return "skus"
}
