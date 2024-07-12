package models

import (
	"time"
)

type TypeIP string

const (
	ALL TypeIP = "all"
	GGA TypeIP = "gga"
	HDT TypeIP = "hdt"
	VTG TypeIP = "vtg"
)

type IPKapal struct {
	IdIpKapal uint      `gorm:"primary_key" json:"id_ip_kapal"`
	CallSign  string    `gorm:"not null;index" json:"call_sign" binding:"required"`
	TypeIP    TypeIP    `gorm:"type:enum('all','gga','hdt','vtg');not null" json:"type_ip" binding:"required"`
	IP        string    `gorm:"type:varchar(16);not null;" json:"ip" binding:"required"`
	Port      uint16    `json:"port;not null;" binding:"required"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Kapal     *Kapal    `gorm:"foreignKey:CallSign;association_foreignkey:CallSign"`
	// Kapal     Kapal     `gorm:"references:CallSign;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

}
