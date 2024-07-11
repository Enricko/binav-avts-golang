package models

import "time"

type Coordinate struct {
	IdCoordinate uint   `gorm:"primary_key" json:"id_ip_kapal"`
	CallSign     string `gorm:"not null;index" json:"call_sign" binding:"required"`
	SeriesID     uint64 `gorm:"not null" json:"series_id" binding:"required"`

	IdCoorGGA *int `gorm:"index;default:null" json:"id_coor_gga"` // Nullable foreign key
	IdCoorHDT *int `gorm:"index;default:null" json:"id_coor_hdt"` // Nullable foreign key
	IdCoorVTG *int `gorm:"index;default:null" json:"id_coor_vtg"` // Nullable foreign key

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Kapal         *Kapal         `gorm:"foreignKey:CallSign;references:CallSign"`
	CoordinateHdt *CoordinateHdt `gorm:"foreignKey:IdCoorHDT;references:IdCoorHDT"`
	CoordinateGga *CoordinateGga `gorm:"foreignKey:IdCoorGGA;references:IdCoorGGA"`
	CoordinateVtg *CoordinateVtg `gorm:"foreignKey:IdCoorVTG;references:IdCoorVTG"`
}
