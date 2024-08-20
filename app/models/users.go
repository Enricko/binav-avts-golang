package models

import "time"

type Level string

const (
	USER  Level = "user"
	ADMIN Level = "admin"
	OWNER Level = "owner"
)

type User struct {
	IdUser   string `gorm:"primary_key" json:"id_user"`
	Name     string `gorm:"varchar(300);not null;" json:"name" binding:"required"`
	Email    string `gorm:"unique_index;not null;" binding:"required" json:"email"`
	Password string `gorm:"varchar(300);not null;" json:"-" binding:"min=6"`
	Level    Level  `gorm:"type:enum('user','admin','owner');not null;" json:"level" binding:"required"`

	ResetOTP       string    `json:"reset_otp"`
	ResetOTPExpiry time.Time `json:"reset_otp_expiry"`
	
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
