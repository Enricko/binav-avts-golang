package models

import (
	"errors"
	"fmt"
	"time"

)

type GpsQuality string

const (
	FixNotValid        GpsQuality = "Fix not valid"
	GpsFix             GpsQuality = "GPS fix"
	DifferentialGpsFix GpsQuality = "Differential GPS fix"
	NotApplicable      GpsQuality = "Not applicable"
	RtkFixed           GpsQuality = "RTK Fixed"
	RtkFloat           GpsQuality = "RTK Float"
	InsDeadReckoning   GpsQuality = "INS Dead reckoning"
)

type VesselRecord struct {
	IdVesselRecord uint64 `gorm:"primary_key" json:"id_ vessel_record"`
	CallSign       string `gorm:"not null;index" json:"call_sign" binding:"required"`
	SeriesID       uint64 `gorm:"not null" json:"series_id" binding:"required"`

	Latitude            string     `gorm:"varchar(255)" json:"latitude" binding:"required"`
	Longitude           string     `gorm:"varchar(255)" json:"longitude" binding:"required"`
	HeadingDegree       float64    `gorm:"varchar(255)" json:"heading_degree" binding:"required"`
	SpeedInKnots        float64    `gorm:"" json:"speed_in_knots" binding:"required"`
	GpsQualityIndicator GpsQuality `gorm:"type:enum('Fix not valid','GPS fix','Differential GPS fix','Not applicable','RTK Fixed','RTK Float','INS Dead reckoning');" json:"gps_quality_indicator"`
	WaterDepth          float64    `gorm:"" json:"water_depth" binding:"required"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Kapal *Kapal `gorm:"foreignKey:CallSign;association_foreignkey:CallSign"`
}

func StringToGpsQuality(value string) (GpsQuality, error) {
	fmt.Println(value)
	switch value {
	case string(FixNotValid):
		return FixNotValid, nil
	case string(GpsFix):
		return GpsFix, nil
	case string(DifferentialGpsFix):
		return DifferentialGpsFix, nil
	case string(NotApplicable):
		return NotApplicable, nil
	case string(RtkFixed):
		return RtkFixed, nil
	case string(RtkFloat):
		return RtkFloat, nil
	case string(InsDeadReckoning):
		return InsDeadReckoning, nil
	default:
		return "", errors.New("invalid GpsQuality value")
	}
}
