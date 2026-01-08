package model

import (
	"zjMall/pkg"

	"gorm.io/gorm"
)

type Brand struct {
	pkg.BaseModel
	Version     int            `gorm:"column:version" json:"version"`
	Name        string         `gorm:"column:name" json:"name"`
	LogoURL     string         `gorm:"column:logo_url" json:"logo_url"`
	Country     string         `gorm:"column:country" json:"country"`
	Description string         `gorm:"column:description" json:"description"`
	FirstLetter string         `gorm:"column:first_letter" json:"first_letter"`
	SortOrder   int            `gorm:"column:sort_order" json:"sort_order"`
	Status      int            `gorm:"column:status" json:"status"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// TableName returns the table name for Brand struct
func (Brand) TableName() string {
	return "brands"
}
