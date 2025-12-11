package pkg

import "time"

// BaseModel 基础模型（所有模型都嵌入此结构体）
type BaseModel struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
