package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"golang-app/app/models"
	"golang-app/database"
)

type IPVesselController struct {
	// Dependent services
}

func NewIPVesselController() *IPVesselController {
	return &IPVesselController{
		// Inject services
	}
}

type IpType string

const (
	NMEA       IpType = "nmea"
	WATERDEPTH IpType = "water_depth"
)

func (r *IPVesselController) GetIPVessels(c *gin.Context) {
	callSign := c.Param("call_sign")

	// Check if the vessel exists
	var vessel models.Kapal
	result := database.DB.Where("call_sign = ?", callSign).First(&vessel)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "No vessel found with the given call sign"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error checking vessel existence"})
		}
		return
	}

	var ipVessels []models.IPKapal
	if err := database.DB.Where("call_sign = ?", callSign).Find(&ipVessels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error fetching IP vessels"})
		return
	}

	// Pagination parameters
	draw, _ := strconv.Atoi(c.DefaultQuery("draw", "1"))
	start, _ := strconv.Atoi(c.DefaultQuery("start", "0"))
	length, _ := strconv.Atoi(c.DefaultQuery("length", "10"))

	// Calculate total and filtered record counts
	totalRecords := len(ipVessels)
	filteredRecords := totalRecords

	// Slice the data for pagination
	end := start + length
	if end > filteredRecords {
		end = filteredRecords
	}
	pagedData := ipVessels[start:end]

	c.JSON(http.StatusOK, gin.H{
		"draw":            draw,
		"recordsTotal":    totalRecords,
		"recordsFiltered": filteredRecords,
		"data":            pagedData,
	})
}

func (r *IPVesselController) GetIPVesselByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var ipVessel models.IPKapal
	result := database.DB.First(&ipVessel, id)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "IP Vessel not found"})
		return
	}

	c.JSON(http.StatusOK, ipVessel)
}

func (r *IPVesselController) InsertIPVessel(c *gin.Context) {
	callSign := c.Param("call_sign")

	// Check if the vessel exists
	var vessel models.Kapal
	result := database.DB.Where("call_sign = ?", callSign).First(&vessel)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "No vessel found with the given call sign"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error checking vessel existence"})
		}
		return
	}

	var input struct {
		Type IpType `form:"type" json:"type" binding:"required"`
		IP   string `form:"ip" json:"ip" binding:"required"`
		Port uint16 `form:"port" json:"port" binding:"required"`
	}
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to bind data"})
		return
	}

	// Convert IpType to models.TypeIP
	var typeIP models.TypeIP
	switch input.Type {
	case NMEA:
		typeIP = models.ALL
	case WATERDEPTH:
		typeIP = models.DEPTH
	default:
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid IP type"})
		return
	}

	ipVessel := models.IPKapal{
		CallSign: callSign,
		TypeIP:   typeIP,
		IP:       input.IP,
		Port:     input.Port,
	}

	if err := database.DB.Create(&ipVessel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create IP Vessel: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "IP Vessel created successfully", "data": ipVessel})
}

func (r *IPVesselController) UpdateIPVessel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}

	// Check if the IP vessel exists
	var existingIPVessel models.IPKapal
	result := database.DB.First(&existingIPVessel, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "IP Vessel not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error checking IP Vessel existence"})
		}
		return
	}

	// Check if the associated vessel exists
	var vessel models.Kapal
	result = database.DB.Where("call_sign = ?", existingIPVessel.CallSign).First(&vessel)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "Associated vessel not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error checking vessel existence"})
		}
		return
	}

	var input struct {
		Type IpType `form:"type" json:"type" binding:"required"`
		IP   string `form:"ip" json:"ip" binding:"required"`
		Port uint16 `form:"port" json:"port" binding:"required"`
	}
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to bind data"})
		return
	}

	// Convert IpType to models.TypeIP
	var typeIP models.TypeIP
	switch input.Type {
	case NMEA:
		typeIP = models.ALL
	case WATERDEPTH:
		typeIP = models.DEPTH
	default:
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid IP type"})
		return
	}

	// Update the IP vessel
	existingIPVessel.TypeIP = typeIP
	existingIPVessel.IP = input.IP
	existingIPVessel.Port = input.Port

	if err := database.DB.Save(&existingIPVessel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update IP Vessel: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "IP Vessel updated successfully", "data": existingIPVessel})
}

func (r *IPVesselController) DeleteIPVessel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}

	// Check if the IP vessel exists
	var ipVessel models.IPKapal
	result := database.DB.First(&ipVessel, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "IP Vessel not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error checking IP Vessel existence"})
		}
		return
	}

	// Optional: Check if the user confirmed the deletion
	var input struct {
		Confirmation string `form:"confirmation" json:"confirmation" binding:"required"`
	}
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Confirmation required"})
		return
	}

	if input.Confirmation != ipVessel.CallSign {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Confirmation does not match Call Sign"})
		return
	}

	// Perform the deletion
	if err := database.DB.Delete(&ipVessel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete IP Vessel: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "IP Vessel deleted successfully"})
}

// Helper function to check if a struct contains a search string
func contains(ipVessel models.IPKapal, search string) bool {
	return strings.Contains(strings.ToLower(ipVessel.CallSign), strings.ToLower(search)) ||
		strings.Contains(strings.ToLower(string(ipVessel.TypeIP)), strings.ToLower(search)) ||
		strings.Contains(strings.ToLower(ipVessel.IP), strings.ToLower(search)) ||
		strings.Contains(strings.ToLower(strconv.Itoa(int(ipVessel.Port))), strings.ToLower(search))
}
