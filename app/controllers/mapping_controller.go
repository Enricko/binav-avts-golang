package controllers

import (
	"golang-app/app/models"
	"golang-app/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MappingController struct {
	// Dependent services
}

func NewMappingController() *MappingController {
	return &MappingController{
		// Inject services
	}
}
func (r *MappingController) GetMappings(c *gin.Context) {
	var mappings []models.Mapping
	if err := database.DB.Preload("User").Find(&mappings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, mappings)
}

func (r *MappingController) GetKMZFile(c *gin.Context) {
	id := c.Param("id")
	var mapping models.Mapping
	if err := database.DB.First(&mapping, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "KMZ file not found"})
		return
	}
	c.Header("Content-Type", "application/vnd.google-earth.kmz")
	c.String(http.StatusOK, mapping.File)
}
