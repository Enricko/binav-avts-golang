package controllers

import (
	"fmt"
	"golang-app/app/models"
	"golang-app/database"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

)

type VesselController struct {
	// Dependent services
}

func NewVesselController() *VesselController {
	return &VesselController{
		// Inject services
	}
}

func (r *VesselController) GetVessel(c *gin.Context) {
	// TODO: Change if want to use another model here
	var data []models.Kapal

	// Fetch users data from the database
	database.DB.Find(&data)

	draw, _ := strconv.Atoi(c.Query("draw"))
	start, _ := strconv.Atoi(c.Query("start"))
	length, _ := strconv.Atoi(c.Query("length"))
	search := c.Query("search[value]") // Get search value from DataTables request

	// Filter data based on search query
	// TODO: Change if want to use another model here
	filteredData := make([]models.Kapal, 0)
	for _, vessel := range data {
		// Iterate over each field of the user struct
		v := reflect.ValueOf(vessel)
		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i).Interface()
			// Check if the field value contains the search query
			if strings.Contains(strings.ToLower(fmt.Sprintf("%v", fieldValue)), strings.ToLower(search)) {
				// If any field contains the search query, add the user to filtered data and break the loop
				filteredData = append(filteredData, vessel)
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
				return filteredData[i].CallSign < filteredData[j].CallSign
			})
		} else {
			sort.Slice(filteredData, func(i, j int) bool {
				return filteredData[i].CallSign > filteredData[j].CallSign
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

func (r *VesselController) InsertVessel(c *gin.Context) {
	var input struct {
		CallSign                    string `form:"call_sign" json:"call_sign" binding:"required"`
		Flag                        string `form:"flag" json:"flag" binding:"required"`
		Kelas                       string `form:"kelas" json:"kelas" binding:"required"`
		Builder                     string `form:"builder" json:"builder" binding:"required"`
		YearBuilt                   *uint  `form:"year_built" json:"year_built" binding:"required"`
		HeadingDirection            *int64 `form:"heading_direction" json:"heading_direction" binding:"required"`
		Calibration                 *int64 `form:"calibration" json:"calibration" binding:"required"`
		WidthM                      *int64 `form:"width_m" json:"width_m" binding:"required"`
		Height                      *int64 `form:"height_m" json:"height_m" binding:"required"`
		TopRange                    *int64 `form:"top_range" json:"top_range" binding:"required"`
		LeftRange                   *int64 `form:"left_range" json:"left_range" binding:"required"`
		HistoryPerSecond            *int64 `form:"history_per_second" json:"history_per_second" binding:"required"`
		MinimumKnotPerLiterGasoline *int64 `form:"minimum_knot_per_liter_gasoline" json:"minimum_knot_per_liter_gasoline" binding:"required"`
		MaximumKnotPerLiterGasoline *int64 `form:"maximum_knot_per_liter_gasoline" json:"maximum_knot_per_liter_gasoline" binding:"required"`
		Status                      bool   `form:"status" json:"status" binding:"required"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to bind data"})
		return
	}

	// Validate that all required fields are present
	if input.YearBuilt == nil || input.HeadingDirection == nil || input.Calibration == nil ||
		input.WidthM == nil || input.Height == nil || input.TopRange == nil || input.LeftRange == nil ||
		input.HistoryPerSecond == nil || input.MinimumKnotPerLiterGasoline == nil || input.MaximumKnotPerLiterGasoline == nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "All fields are required"})
		return
	}

	// Handle image_map upload
	imageMapFile, err := c.FormFile("image_map")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to upload image map"})
		return
	}

	// Handle image upload
	imageFile, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to upload vessel image"})
		return
	}

	// Create directory if it doesn't exist
	dir := "public/upload/assets/image/vessel"
	dir2 := "public/upload/assets/image/vessel_map"
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create directory"})
		return
	}
	if err := os.MkdirAll(dir2, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create directory"})
		return
	}

	// Save the image map file with a unique name
	imageMapFilename := time.Now().Format("2006-01-02 15_04_05") + input.CallSign + filepath.Ext(imageMapFile.Filename)
	imageMapPath := filepath.Join(dir2, imageMapFilename)
	if err := c.SaveUploadedFile(imageMapFile, imageMapPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to save image map"})
		return
	}

	// Save the vessel image file with a unique name
	imageFilename := time.Now().Format("2006-01-02 15_04_05") + input.CallSign + filepath.Ext(imageFile.Filename)
	imagePath := filepath.Join(dir, imageFilename)
	if err := c.SaveUploadedFile(imageFile, imagePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to save vessel image"})
		return
	}

	vessel := models.Kapal{
		CallSign:                    input.CallSign,
		Flag:                        input.Flag,
		Kelas:                       input.Kelas,
		Builder:                     input.Builder,
		YearBuilt:                   *input.YearBuilt,
		HeadingDirection:            *input.HeadingDirection,
		Calibration:                 *input.Calibration,
		WidthM:                      *input.WidthM,
		Height:                      *input.Height,
		TopRange:                    *input.TopRange,
		LeftRange:                   *input.LeftRange,
		ImageMap:                    imageMapFilename,
		Image:                       imageFilename,
		HistoryPerSecond:            *input.HistoryPerSecond,
		MinimumKnotPerLiterGasoline: *input.MinimumKnotPerLiterGasoline,
		MaximumKnotPerLiterGasoline: *input.MaximumKnotPerLiterGasoline,
		Status:                      input.Status,
	}

	if err := database.DB.Create(&vessel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vessel created successfully", "data": vessel})
}
