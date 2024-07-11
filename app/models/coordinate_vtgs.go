package models

import "time"

// ModeIndicator represents the possible values for mode indicators.
type ModeIndicator string

const (
    AutonomousMode          ModeIndicator = "Autonomous mode"
    DifferentialMode        ModeIndicator = "Differential mode"
    EstimatedMode           ModeIndicator = "Estimated (dead reckoning) mode"
    ManualInputMode         ModeIndicator = "Manual Input mode"
    SimulatorMode           ModeIndicator = "Simulator mode"
    DataNotValidMode        ModeIndicator = "Data not valid"
)


// VtgCoordinate represents the structure of the 'vtg_coordinates' table in the database.
type CoordinateVtg struct {

	IdCoorVTG     int       `gorm:"primary_key" json:"id_coor_vtg"`
	CallSign      string    `gorm:"not null;index" json:"call_sign" binding:"required"`
	MessageID     string    `gorm:"not null" json:"message_id" binding:"required"`
    TrackDegreeTrue     float32       `gorm:"not null" json:"track_degree_true" binding:"required"`
    TrueNorth           string        `gorm:"varchar(1);not null" json:"track_degree_true" binding:"required"`
    TrackDegreeMagnetic float32       `gorm:"not null" json:"track_degree_magnetic" binding:"required"`
    MagneticNorth       string        `gorm:"varchar(1);not null" json:"track_degree_magnetic" binding:"required"`
    SpeedInKnots        float32       `gorm:"not null" json:"speed_in_knots" binding:"required"`
    MeasuredKnots       string        `gorm:"varchar(1);not null" json:"measured_knots" binding:"required"`
    Kph                 float32       `gorm:"not null" json:"kph" binding:"required"`
    MeasuredKph         string        `gorm:"varchar(1);not null" json:"measured_kph" binding:"required"`
    ModeIndicator       ModeIndicator `gorm:"type:enum('Autonomous mode','Differential mode','Estimated (dead reckoning) mode','Manual Input mode','Simulator mode','Data not valid');" json:"mode_indicator"`

	Checksum      string    `gorm:"not null" json:"checksum" binding:"required"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	
	Kapal *Kapal `gorm:"foreignKey:CallSign;references:CallSign"`

}