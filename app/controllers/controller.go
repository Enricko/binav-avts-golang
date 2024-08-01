package controllers

import (
	"golang-app/app/models"
	"golang-app/database"
	"net/http"

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
	if err := database.DB.Where("call_sign = ?", callSign).First(&kapal).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kapal not found"})
		return
	}
	result := database.DB.Where("call_sign = ?", callSign).Find(&records)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"call_sign": callSign,
		"kapal":     kapal,
		"records":   records,
	})
}
