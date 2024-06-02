package models

import "time"

type User struct {
	IdUser    string    `gorm:"primary_key" binding:"required" json:"id_user"`
	Name      string    `gorm:"varchar(300);not null;" json:"name" binding:"required"`
	Password  string    `gorm:"varchar(300);not null;" json:"-" binding:"required,min=6"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
