package models

import "time"

type CoordinateHdt struct {
	IdCoorHDT     int       `gorm:"primary_key" json:"id_coor_hdt"`
	CallSign      string    `gorm:"not null;index" json:"call_sign" binding:"required"`
	MessageID     string    `gorm:"not null" json:"message_id" binding:"required"`
	HeadingDegree float32   `gorm:"not null" json:"heading_degree" binding:"required"`
	Checksum      string    `gorm:"not null" json:"checksum" binding:"required"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	Kapal *Kapal `gorm:"foreignKey:CallSign;association_foreignkey:CallSign"`
}
