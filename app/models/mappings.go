package models

import "time"

type Mapping struct {
	IdMapping uint `gorm:"primary_key" json:"id_mapping"`

	Name       string    `gorm:"varchar(255);not null;" json:"name" binding:"required"`
	File       string    `gorm:"type:TEXT;not null;" json:"file" binding:"required"`
	Status     bool      `gorm:"not_null" json:"status" binding:"required"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

}