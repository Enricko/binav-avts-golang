package models

import "time"

type Size string

const (
	Small      Size = "small"
	Medium     Size = "medium"
	Large      Size = "large"
	ExtraLarge Size = "extra_large"
)

type Kapal struct {
	CallSign         string    `gorm:"primary_key" binding:"required" json:"call_sign"`
	Status           bool      `json:"status;not null" binding:"required"`
	Flag             string    `gorm:"varchar(300);not null;" json:"flag" binding:"required"`
	Kelas            string    `gorm:"varchar(300);not null;" json:"kelas" binding:"required"`
	Builder          string    `gorm:"varchar(300);not null;" json:"builder" binding:"required"`
	YearBuilt        uint      `gorm:";not null;" json:"year_built" binding:"required"`
	HeadingDirection int64     `gorm:"varchar(300);not null;" json:"heading_direction" binding:"required"`
	Size             Size      `gorm:"not null;type:enum('small','medium','large','extra_large')" json:"size" binding:"required"`
	XmlFile          string    `gorm:"type:TEXT;not null;" json:"XmlFile" binding:"required"`
	Image            string    `gorm:"type:TEXT;not null;" json:"Image" binding:"required"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
