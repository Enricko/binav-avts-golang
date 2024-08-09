package controllers

import (
	"golang-app/app/models"
	"golang-app/database"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"

)

type Controller struct {
	// Dependent services
}

func NewController() *Controller {
	return &Controller{}
}

func (r *Controller) Index(c *gin.Context) {
	// Data to pass to the index.html template
	data := gin.H{
		"title":     "Welcome Administrator",
		"csrfToken": csrf.GetToken(c),
	}
	// Render the index.html template with data
	c.HTML(http.StatusOK, "index.html", data)
}

func (r *Controller) Login(c *gin.Context) {
	// Data to pass to the index.html template
	data := gin.H{
		"title":     "Login Page",
		"csrfToken": csrf.GetToken(c),
	}
	// Render the index.html template with data
	c.HTML(http.StatusOK, "login.html", data)
}
func (r *Controller) GetVesselRecords(c *gin.Context) {
	callSign := c.Param("call_sign")
	var kapal models.Kapal
	var records []models.VesselRecord

	// Parse start and end datetime from query parameters
	start := c.Query("start")
	end := c.Query("end")

	// Set default start to 3 days ago and end to now if not provided
	if start == "" || end == "" {
		now := time.Now()
		defaultEnd := now.Format("2006-01-02 15:04:05")
		defaultStart := now.AddDate(0, 0, -3).Format("2006-01-02 15:04:05")
		if start == "" {
			start = defaultStart
		}
		if end == "" {
			end = defaultEnd
		}
	}

	// Find the Kapal with the given call sign
	if err := database.DB.Where("call_sign = ?", callSign).First(&kapal).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kapal not found"})
		return
	}

	// Initialize the query for fetching records
	query := database.DB.Where("call_sign = ?", callSign)

	// Apply datetime range filter
	query = query.Where("created_at BETWEEN ? AND ?", start, end)

	// Execute the query to find the records
	result := query.Find(&records)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// Respond with the records and additional information
	c.JSON(http.StatusOK, gin.H{
		"call_sign":    callSign,
		"kapal":        kapal,
		"records":      records,
		"total_record": len(records),
	})
}
