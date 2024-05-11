package models

import (
    "github.com/jinzhu/gorm"

)

type User struct {
    gorm.Model
    Username string `gorm:"unique_index"`
    Email    string `gorm:"unique_index"`
}