package controllers

import (
	"fmt"
	"golang-app/app/models"
	"golang-app/database"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"

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
func (r *MappingController) GetAllMapping(c *gin.Context) {
	var data []models.Mapping

	// Preload User data while querying Mapping records
	result := database.DB.Preload("User").Find(&data)
	if result.Error != nil {
		// Log or handle the error
		fmt.Println("Error loading associated user data:", result.Error)
		// Handle the error, e.g., return an error response
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// Debug: Print the data
	for _, mapping := range data {
		fmt.Printf("%+v\n", mapping)
	}

	draw, _ := strconv.Atoi(c.Query("draw"))
	start, _ := strconv.Atoi(c.Query("start"))
	length, _ := strconv.Atoi(c.Query("length"))
	search := c.Query("search[value]")

	filteredData := make([]models.Mapping, 0)
	for _, mapping := range data {
		v := reflect.ValueOf(mapping)
		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i).Interface()
			if strings.Contains(strings.ToLower(fmt.Sprintf("%v", fieldValue)), strings.ToLower(search)) {
				filteredData = append(filteredData, mapping)
				break
			}
		}
	}

	totalRecords := len(filteredData)

	// Sort data based on order specified by DataTables
	orderColumnIndex, _ := strconv.Atoi(c.Query("order[0][column]"))
	orderDirection := c.Query("order[0][dir]")

	switch orderColumnIndex {
	case 0: // Assuming sorting is based on ID column
		if orderDirection == "asc" {
			sort.Slice(filteredData, func(i, j int) bool {
				return filteredData[i].IdMapping < filteredData[j].IdMapping
			})
		} else {
			sort.Slice(filteredData, func(i, j int) bool {
				return filteredData[i].IdMapping > filteredData[j].IdMapping
			})
		}
		// Add cases for other columns if needed
	}

	// Slice the data to return only the portion needed for the current page
	end := start + length
	if end > totalRecords {
		end = totalRecords
	}
	pagedData := filteredData[start:end]

	// Send JSON response to DataTables
	c.JSON(http.StatusOK, gin.H{
		"draw":            draw,
		"recordsTotal":    len(data),
		"recordsFiltered": totalRecords,
		"data":            pagedData,
	})
}
