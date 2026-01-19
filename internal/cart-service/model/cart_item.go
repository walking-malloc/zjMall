package model

import (
	"fmt"
	"time"
	"zjMall/pkg"
)

// CartItem 购物车项模型
// 对应数据库表：cart_items
type CartItem struct {
	pkg.BaseModel

	// 关联信息
	UserID string `gorm:"type:varchar(26);not null;comment:用户ID" json:"user_id"`

	// 商品信息
	ProductID    string `gorm:"type:varchar(26);not null;comment:商品ID（SPU ID）" json:"product_id"`
	SKUID        string `gorm:"type:varchar(26);not null;comment:SKU ID" json:"sku_id"`
	ProductTitle string `gorm:"type:varchar(200);not null;comment:商品标题" json:"product_title"`
	ProductImage string `gorm:"type:varchar(255);comment:商品主图" json:"product_image"`
	SKUName      string `gorm:"type:varchar(100);comment:SKU 名称（规格描述）" json:"sku_name"`

	// 价格信息（使用 DECIMAL 类型，在数据库中存储为 DECIMAL(10,2)）
	Price        float64 `gorm:"type:decimal(10,2);not null;comment:单价（加购时的价格快照）" json:"price"`
	CurrentPrice float64 `gorm:"type:decimal(10,2);not null;comment:当前价格（实时查询）" json:"current_price"`

	// 数量信息
	Quantity int32 `gorm:"type:int;not null;default:1;comment:数量" json:"quantity"`
	Stock    int32 `gorm:"type:int;not null;default:0;comment:当前库存" json:"stock"`

	// 状态信息
	IsValid       bool   `gorm:"type:tinyint(1);default:1;comment:是否有效：0-无效，1-有效" json:"is_valid"`
	InvalidReason string `gorm:"type:varchar(100);comment:失效原因" json:"invalid_reason,omitempty"`
}

// TableName 指定表名
func (CartItem) TableName() string {
	return "cart_items"
}

// PriceString 返回价格的字符串形式（用于 Proto 转换）
func (c *CartItem) PriceString() string {
	return formatPrice(c.Price)
}

// CurrentPriceString 返回当前价格的字符串形式
func (c *CartItem) CurrentPriceString() string {
	return formatPrice(c.CurrentPrice)
}

// formatPrice 格式化价格为字符串（保留2位小数）
func formatPrice(price float64) string {
	return formatFloat(price, 2)
}

// formatFloat 格式化浮点数为字符串
func formatFloat(f float64, precision int) string {
	format := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(format, f)
}

// ToProto 转换为 Proto 格式
func (c *CartItem) ToProto() map[string]interface{} {
	return map[string]interface{}{
		"id":             c.ID,
		"user_id":        c.UserID,
		"product_id":     c.ProductID,
		"sku_id":         c.SKUID,
		"product_title":  c.ProductTitle,
		"product_image":  c.ProductImage,
		"sku_name":       c.SKUName,
		"price":          c.PriceString(),
		"current_price":  c.CurrentPriceString(),
		"quantity":       c.Quantity,
		"stock":          c.Stock,
		"is_valid":       c.IsValid,
		"invalid_reason": c.InvalidReason,
		"created_at":     c.CreatedAt,
		"updated_at":     c.UpdatedAt,
	}
}

// FromProto 从 Proto 格式创建
func CartItemFromProto(data map[string]interface{}) *CartItem {
	item := &CartItem{}
	if id, ok := data["id"].(string); ok {
		item.ID = id
	}
	if userID, ok := data["user_id"].(string); ok {
		item.UserID = userID
	}
	if productID, ok := data["product_id"].(string); ok {
		item.ProductID = productID
	}
	if skuID, ok := data["sku_id"].(string); ok {
		item.SKUID = skuID
	}
	if title, ok := data["product_title"].(string); ok {
		item.ProductTitle = title
	}
	if image, ok := data["product_image"].(string); ok {
		item.ProductImage = image
	}
	if skuName, ok := data["sku_name"].(string); ok {
		item.SKUName = skuName
	}
	// 价格从字符串转换为 float64
	if priceStr, ok := data["price"].(string); ok {
		if price, err := parsePrice(priceStr); err == nil {
			item.Price = price
		}
	}
	if currentPriceStr, ok := data["current_price"].(string); ok {
		if currentPrice, err := parsePrice(currentPriceStr); err == nil {
			item.CurrentPrice = currentPrice
		}
	}
	if quantity, ok := data["quantity"].(int32); ok {
		item.Quantity = quantity
	}
	if stock, ok := data["stock"].(int32); ok {
		item.Stock = stock
	}
	if isValid, ok := data["is_valid"].(bool); ok {
		item.IsValid = isValid
	}
	if reason, ok := data["invalid_reason"].(string); ok {
		item.InvalidReason = reason
	}
	if createdAt, ok := data["created_at"].(time.Time); ok {
		item.CreatedAt = createdAt
	}
	if updatedAt, ok := data["updated_at"].(time.Time); ok {
		item.UpdatedAt = updatedAt
	}
	return item
}

// parsePrice 解析价格字符串为 float64
func parsePrice(priceStr string) (float64, error) {
	var price float64
	_, err := fmt.Sscanf(priceStr, "%f", &price)
	return price, err
}
