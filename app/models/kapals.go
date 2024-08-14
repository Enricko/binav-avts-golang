package models

import "time"

type Kapal struct {
	CallSign         string    `gorm:"primary_key" binding:"required" json:"call_sign"`
	Flag             string    `gorm:"varchar(300);not null;" json:"flag" binding:"required"`
	Kelas            string    `gorm:"varchar(300);not null;" json:"kelas" binding:"required"`
	Builder          string    `gorm:"varchar(300);not null;" json:"builder" binding:"required"`
	YearBuilt        uint      `gorm:";not null;" json:"year_built" binding:"required"`
	HeadingDirection int64     `gorm:"varchar(300);not null;" json:"heading_direction" binding:"required"`
	Calibration      int64     `gorm:"varchar(300);not null;" json:"calibration" binding:"required"`
	WidthM           int64     `gorm:"not null;" json:"width_m" binding:"required"`
	Height           int64     `gorm:"not null;" json:"height_m" binding:"required"`
	TopRange         int64     `gorm:"not null;" json:"top_range" binding:"required"`
	LeftRange        int64     `gorm:"not null;" json:"left_range" binding:"required"`
	ImageMap         string    `gorm:"type:TEXT;not null;" json:"image_map" binding:"required"`
	Image            string    `gorm:"type:TEXT;not null;" json:"image" binding:"required"`
	HistoryPerSecond int64     `gorm:"default:'1';" json:"history_per_second" binding:"required"`
	MinimumKnotPerLiterGasoline int64 `gorm:"not null;" json:"minimum_knot_per_liter_gasoline" binding:"required"`
	MaximumKnotPerLiterGasoline int64 `gorm:"not null;" json:"maximum_knot_per_liter_gasoline" binding:"required"`
	Status           bool      `gorm:"not null" json:"status" binding:"required"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
