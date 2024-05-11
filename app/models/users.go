package models

import (
    "github.com/jinzhu/gorm"

)

type User struct {
    Username string `gorm:"unique_index"`
    Email    string `gorm:"unique_index"`
    gorm.Model
}