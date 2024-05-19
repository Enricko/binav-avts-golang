package models

import (
    "github.com/jinzhu/gorm"

)

type User struct {
    Username string `gorm:"unique_index" binding:"required"`
    Email    string `gorm:"unique_index" binding:"required"`
    gorm.Model
}