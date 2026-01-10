package model

import (
	"time"

	"zjMall/pkg"

	"gorm.io/gorm"
)

// Product 商品模型（SPU）
// 对应数据库表：products
type Product struct {
	pkg.BaseModel

	// 关联信息
	CategoryID string `gorm:"type:varchar(26);not null;comment:所属类目ID" json:"category_id"`
	BrandID    string `gorm:"type:varchar(26);comment:品牌ID" json:"brand_id,omitempty"`

	// 基本信息
	Title       string `gorm:"type:varchar(200);not null;comment:商品标题" json:"title"`
	Subtitle    string `gorm:"type:varchar(200);comment:商品副标题/卖点" json:"subtitle,omitempty"`
	MainImage   string `gorm:"type:varchar(255);not null;comment:主图URL" json:"main_image"`
	Images      string `gorm:"type:text;comment:轮播图URL列表（JSON数组）" json:"images,omitempty"`
	Description string `gorm:"type:text;comment:商品详情（富文本）" json:"description,omitempty"`

	// 状态与上下架

	Status       int8       `gorm:"type:tinyint;not null;default:1;comment:状态：1-草稿，2-待审核，3-已上架，4-已下架，5-已删除" json:"status"`
	OnShelfTime  *time.Time `gorm:"type:timestamp;comment:上架时间" json:"on_shelf_time,omitempty"`
	OffShelfTime *time.Time `gorm:"type:timestamp;comment:下架时间" json:"off_shelf_time,omitempty"`

	// 软删除
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// TableName 指定表名
func (Product) TableName() string {
	return "products"
}
