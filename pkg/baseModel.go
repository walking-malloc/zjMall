package pkg

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型（所有模型都嵌入此结构体）
// ID 使用 ULID（字符串类型），相比自增 ID 的优势：
// 1. 不暴露业务信息（不会暴露用户数量）
// 2. 分布式友好（不需要中心化 ID 生成器）
// 3. 可排序（基于时间戳，方便查询最新数据）
// 4. 全局唯一（UUID 级别）
type BaseModel struct {
	ID        string    `gorm:"type:varchar(26);primaryKey;default:''" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate GORM 钩子：在创建记录前生成 ULID
func (m *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = GenerateULID()
	}
	return nil
}
