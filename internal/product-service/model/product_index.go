package model

// ProductIndex 商品搜索索引结构
type ProductIndex struct {
	ID              string      `json:"id"`
	Title           string      `json:"title"`       // 商品标题
	Subtitle        string      `json:"subtitle"`    // 副标题
	Description     string      `json:"description"` // 描述
	CategoryID      string      `json:"category_id"`
	CategoryName    string      `json:"category_name"` // 类目名称
	BrandID         string      `json:"brand_id"`
	BrandName       string      `json:"brand_name"`              // 品牌名称
	Tags            []string    `json:"tags"`                    // 标签列表
	SKUs            []*SKUIndex `json:"skus"`                    // SKU列表
	AttributeValues []string    `json:"attribute_values"`        // 属性列表
	Status          int8        `json:"status"`                  // 状态：3-已上架
	OnShelfTime     *string     `json:"on_shelf_time,omitempty"` // 上架时间，可能为空
	CreatedAt       string      `json:"created_at"`
	UpdatedAt       string      `json:"updated_at"`
}

type SKUIndex struct {
	SKUName string  `json:"sku_name"` // 如：红色、XL
	Price   float64 `json:"price"`    // 价格
}
