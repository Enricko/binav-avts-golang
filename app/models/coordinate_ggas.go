package models

import (
	"time"

)

type GpsQuality string

const (
    FixNotValid                             GpsQuality = "Fix not valid"
    GpsFix                                  GpsQuality = "GPS fix"
    DifferentialGpsFix                      GpsQuality = "Differential GPS fix (DGNSS), SBAS, OmniSTAR VBS, Beacon, RTX in GVBS mode"
    NotApplicable                           GpsQuality = "Not applicable"
    RtkFixed                                GpsQuality = "RTK Fixed, xFill"
    RtkFloat                                GpsQuality = "RTK Float, OmniSTAR XP/HP, Location RTK, RTX"
    InsDeadReckoning                        GpsQuality = "INS Dead reckoning"
)
// GgaCoordinate represents the structure of the 'gga_coordinates' table in the database.
type CoordinateGga struct {
	IdCoorGGA           int       `gorm:"primary_key" json:"id_coor_gga"`
	CallSign            string    `gorm:"not null;index" json:"call_sign" binding:"required"`
	MessageID           string    `gorm:"not null" json:"message_id" binding:"required"`
	UtcPosition         float32   `gorm:"not null" json:"utc_position" binding:"required"`
	Latitude            float32   `gorm:"not null" json:"latitude" binding:"required"`
	DirectionLatitude   string    `gorm:"varchar(1);not null" json:"direction_latitude" binding:"required"`
	Longitude           float32   `gorm:"not null" json:"longitude" binding:"required"`
	DirectionLongitude  string    `gorm:"varchar(1);not null" json:"direction_longitude" binding:"required"`
	GpsQualityIndicator GpsQuality    `gorm:"type:enum('Fix not valid','GPS fix','Differential GPS fix (DGNSS), SBAS, OmniSTAR VBS, Beacon, RTX in GVBS mode','Not applicable','RTK Fixed, xFill','RTK Float, OmniSTAR XP/HP, Location RTK, RTX','INS Dead reckoning');not null" json:"gps_quality_indicator" binding:"required"`
	NumberSv            int       `gorm:"not null" json:"number_sv" binding:"required"`
	Hdop                float32   `gorm:"not null" json:"hdop" binding:"required"`
	OrthometricHeight   float32   `gorm:"not null" json:"orthometric_height" binding:"required"`
	UnitMeasure         string    `gorm:"varchar(255);not null" json:"unit_measure" binding:"required"`
	GeoidSeparation     float32   `gorm:"not null" json:"geoid_separation" binding:"required"`
	GeoidMeasure        string    `gorm:"varchar(255);not null" json:"geoid_measure" binding:"required"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Kapal *Kapal `gorm:"foreignKey:CallSign;association_foreignkey:CallSign"`
}
