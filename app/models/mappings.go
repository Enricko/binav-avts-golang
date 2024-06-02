package models

import "time"

type Mapping struct {
	IdMapping uint `gorm:"primary_key" json:"id_mapping"`

	IdUser string `gorm:"not null;index" json:"id_user" binding:"required"`
	Name   string `gorm:"varchar(255);not null;" json:"name" binding:"required"`
	File   string `gorm:type:TEXT;not null;" json:"file" binding:"required"`

	Status    bool      `gorm:"not_null" json:"status" binding:"required"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User *User `gorm:"foreignKey:IdUser;references:IdUser"`
}
