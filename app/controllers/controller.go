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
	var records []models.VesselRecord
	result := database.DB.Find(&records)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, records)
}
