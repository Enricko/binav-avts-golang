package models

type Coordinate struct {
	IdCoordinate uint   `gorm:"primary_key" json:"id_ip_kapal"`
	CallSign     string `gorm:"not null;index" json:"call_sign" binding:"required"`
	SeriesID     uint64 `gorm:"not null" json:"series_id" binding:"required"`

	IdCoorGGA int `gorm:"index" json:"id_coor_gga" binding:"required"`
	IdCoorHDT int `gorm:"index" json:"id_coor_hdt" binding:"required"`
	IdCoorVTG int `gorm:"index" json:"id_coor_vtg" binding:"required"`

	Kapal         *Kapal         `gorm:"foreignKey:CallSign;references:CallSign"`
	CoordinateGga *CoordinateGga `gorm:"foreignKey:IdCoorGGA;references:IdCoorGGA"`
	CoordinateHdt *CoordinateHdt `gorm:"foreignKey:IdCoorHDT;references:IdCoorHDT"`
	CoordinateVtg *CoordinateVtg `gorm:"foreignKey:IdCoorVTG;references:IdCoorVTG"`
	// Kapal     Kapal     `gorm:"references:CallSign;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

}
