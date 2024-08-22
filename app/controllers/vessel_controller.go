package controllers

import (
	"encoding/json"
	"fmt"
	"golang-app/app/models"
	"golang-app/database"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

func (r *VesselController) GetVesselByCallSign(c *gin.Context) {
	callSign := c.Param("call_sign")

	var vessel models.Kapal
	result := database.DB.Where("call_sign = ?", callSign).First(&vessel)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vessel not found"})
		return
	}

	c.JSON(http.StatusOK, vessel)
}

func (r *VesselController) InsertVessel(c *gin.Context) {
	var input struct {
		CallSign                    string   `form:"call_sign" json:"call_sign" binding:"required"`
		Flag                        string   `form:"flag" json:"flag" binding:"required"`
		Kelas                       string   `form:"kelas" json:"kelas" binding:"required"`
		Builder                     string   `form:"builder" json:"builder" binding:"required"`
		YearBuilt                   *uint    `form:"year_built" json:"year_built" binding:"required"`
		HeadingDirection            *int64   `form:"heading_direction" json:"heading_direction" binding:"required"`
		Calibration                 *int64   `form:"calibration" json:"calibration" binding:"required"`
		WidthM                      *int64   `form:"width_m" json:"width_m" binding:"required"`
		Height                      *int64   `form:"height_m" json:"height_m" binding:"required"`
		TopRange                    *int64   `form:"top_range" json:"top_range" binding:"required"`
		LeftRange                   *int64   `form:"left_range" json:"left_range" binding:"required"`
		HistoryPerSecond            *int64   `form:"history_per_second" json:"history_per_second" binding:"required"`
		MinimumKnotPerLiterGasoline *float64 `form:"minimum_knot_per_liter_gasoline" json:"minimum_knot_per_liter_gasoline" binding:"required"`
		MaximumKnotPerLiterGasoline *float64 `form:"maximum_knot_per_liter_gasoline" json:"maximum_knot_per_liter_gasoline" binding:"required"`
		RecordStatus                bool     `form:"record_status" json:"record_status" binding:"required"`
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
		RecordStatus:                input.RecordStatus,
	}

	if err := database.DB.Create(&vessel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vessel created successfully", "data": vessel})
}
func (r *VesselController) UpdateVessel(c *gin.Context) {
	currentCallSign := c.Param("call_sign")

	var vessel models.Kapal
	if err := database.DB.Where("call_sign = ?", currentCallSign).First(&vessel).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Vessel not found"})
		return
	}

	var input struct {
		CallSign                    string   `form:"call_sign" json:"call_sign"`
		Flag                        string   `form:"flag" json:"flag"`
		Kelas                       string   `form:"kelas" json:"kelas"`
		Builder                     string   `form:"builder" json:"builder"`
		YearBuilt                   *uint    `form:"year_built" json:"year_built"`
		HeadingDirection            *int64   `form:"heading_direction" json:"heading_direction"`
		Calibration                 *int64   `form:"calibration" json:"calibration"`
		WidthM                      *int64   `form:"width_m" json:"width_m"`
		Height                      *int64   `form:"height_m" json:"height_m"`
		TopRange                    *int64   `form:"top_range" json:"top_range"`
		LeftRange                   *int64   `form:"left_range" json:"left_range"`
		HistoryPerSecond            *int64   `form:"history_per_second" json:"history_per_second"`
		MinimumKnotPerLiterGasoline *float64 `form:"minimum_knot_per_liter_gasoline" json:"minimum_knot_per_liter_gasoline"`
		MaximumKnotPerLiterGasoline *float64 `form:"maximum_knot_per_liter_gasoline" json:"maximum_knot_per_liter_gasoline"`
		RecordStatus                string   `form:"record_status" json:"record_status"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to bind data"})
		return
	}

	tx := database.DB.Begin()

	// Check if the call sign is being changed
	if input.CallSign != "" && input.CallSign != currentCallSign {
		// Check if the new call sign already exists
		var existingVessel models.Kapal
		if err := tx.Where("call_sign = ? AND call_sign != ?", input.CallSign, currentCallSign).First(&existingVessel).Error; err == nil {
			tx.Rollback()
			c.JSON(http.StatusConflict, gin.H{"message": "A vessel with this call sign already exists"})
			return
		}
		vessel.CallSign = input.CallSign
	}

	// Update other fields
	if input.Flag != "" {
		vessel.Flag = input.Flag
	}
	if input.Kelas != "" {
		vessel.Kelas = input.Kelas
	}
	if input.Builder != "" {
		vessel.Builder = input.Builder
	}
	if input.YearBuilt != nil {
		vessel.YearBuilt = *input.YearBuilt
	}
	if input.HeadingDirection != nil {
		vessel.HeadingDirection = *input.HeadingDirection
	}
	if input.Calibration != nil {
		vessel.Calibration = *input.Calibration
	}
	if input.WidthM != nil {
		vessel.WidthM = *input.WidthM
	}
	if input.Height != nil {
		vessel.Height = *input.Height
	}
	if input.TopRange != nil {
		vessel.TopRange = *input.TopRange
	}
	if input.LeftRange != nil {
		vessel.LeftRange = *input.LeftRange
	}
	if input.HistoryPerSecond != nil {
		vessel.HistoryPerSecond = *input.HistoryPerSecond
	}
	if input.MinimumKnotPerLiterGasoline != nil {
		vessel.MinimumKnotPerLiterGasoline = *input.MinimumKnotPerLiterGasoline
	}
	if input.MaximumKnotPerLiterGasoline != nil {
		vessel.MaximumKnotPerLiterGasoline = *input.MaximumKnotPerLiterGasoline
	}
	if input.RecordStatus != "" {
		recordStatus, err := strconv.ParseBool(input.RecordStatus)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid record status value"})
			return
		}
		vessel.RecordStatus = recordStatus
	}

	// Handle image_map upload
	imageMapFile, err := c.FormFile("image_map")
	if err == nil {
		dir := "public/upload/assets/image/vessel_map"
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create directory"})
			return
		}

		// Remove old image_map file if it exists
		if vessel.ImageMap != "" {
			oldImageMapPath := filepath.Join(dir, vessel.ImageMap)
			if err := os.Remove(oldImageMapPath); err != nil {
				// Log the error but don't stop the update process
				fmt.Printf("Failed to remove old image map: %v\n", err)
			}
		}

		imageMapFilename := time.Now().Format("2006-01-02 15_04_05") + vessel.CallSign + filepath.Ext(imageMapFile.Filename)
		imageMapPath := filepath.Join(dir, imageMapFilename)
		if err := c.SaveUploadedFile(imageMapFile, imageMapPath); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to save image map"})
			return
		}

		vessel.ImageMap = imageMapFilename
	}

	// Handle image upload
	imageFile, err := c.FormFile("image")
	if err == nil {
		dir := "public/upload/assets/image/vessel"
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create directory"})
			return
		}

		// Remove old image file if it exists
		if vessel.Image != "" {
			oldImagePath := filepath.Join(dir, vessel.Image)
			if err := os.Remove(oldImagePath); err != nil {
				// Log the error but don't stop the update process
				fmt.Printf("Failed to remove old vessel image: %v\n", err)
			}
		}

		imageFilename := time.Now().Format("2006-01-02 15_04_05") + vessel.CallSign + filepath.Ext(imageFile.Filename)
		imagePath := filepath.Join(dir, imageFilename)
		if err := c.SaveUploadedFile(imageFile, imagePath); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to save vessel image"})
			return
		}

		vessel.Image = imageFilename
	}

	// Update the vessel in the database
	if err := tx.Model(&vessel).Where("call_sign = ?", currentCallSign).Save(vessel).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update vessel"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Vessel updated successfully", "data": vessel})
}

func (r *VesselController) DeleteVessel(c *gin.Context) {
	callSign := c.Param("call_sign")

	var input struct {
		ConfirmationName string `form:"confirmationName" json:"confirmationName" binding:"required"`
	}
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "error": err.Error()})
		return
	}

	// Check if the confirmation name matches the actual vessel call sign
	if input.ConfirmationName != callSign {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Confirmation name does not match the vessel call sign"})
		return
	}

	tx := database.DB.Begin()

	// Find the vessel
	var vessel models.Kapal
	if err := tx.Where("call_sign = ?", callSign).First(&vessel).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "Vessel not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error finding vessel"})
		}
		return
	}

	// Delete the vessel record
	if err := tx.Delete(&vessel).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete vessel record"})
		return
	}

	// Delete associated image files
	if vessel.ImageMap != "" {
		imageMapPath := filepath.Join("public/upload/assets/image/vessel_map", vessel.ImageMap)
		if err := os.Remove(imageMapPath); err != nil {
			// Log the error but don't stop the deletion process
			fmt.Printf("Failed to delete image map file: %v\n", err)
		}
	}

	if vessel.Image != "" {
		imagePath := filepath.Join("public/upload/assets/image/vessel", vessel.Image)
		if err := os.Remove(imagePath); err != nil {
			// Log the error but don't stop the deletion process
			fmt.Printf("Failed to delete vessel image file: %v\n", err)
		}
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Vessel deleted successfully"})
}

func (r *VesselController) GetVesselRecords(c *gin.Context) {
	callSign := c.Param("call_sign")
	var kapal models.Kapal

	// Parse start and end datetime from query parameters
	start := c.DefaultQuery("start", time.Now().AddDate(0, 0, -3).Format("2006-01-02 15:04:05"))
	end := c.DefaultQuery("end", time.Now().Format("2006-01-02 15:04:05"))

	// Find the Kapal with the given call sign
	if err := database.DB.Where("call_sign = ?", callSign).First(&kapal).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kapal not found"})
		return
	}

	// Initialize the query for fetching records
	query := database.DB.Where("call_sign = ? AND created_at BETWEEN ? AND ?", callSign, start, end)

	// Count total records
	var totalRecords int64
	if err := query.Model(&models.VesselRecord{}).Count(&totalRecords).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set up streaming response
	c.Header("Content-Type", "application/json")
	c.Header("Transfer-Encoding", "chunked")

	// Start the JSON response
	c.Writer.Write([]byte(fmt.Sprintf(`{"call_sign":"%s","kapal":%s,"total_record":%d,"records":[`,
		callSign, toJSON(kapal), totalRecords)))

	// Stream records
	rows, err := query.Model(&models.VesselRecord{}).Rows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	isFirstRecord := true
	for rows.Next() {
		var record models.VesselRecord
		if err := database.DB.ScanRows(rows, &record); err != nil {
			// Log the error and continue
			log.Printf("Error scanning row: %v", err)
			continue
		}

		if !isFirstRecord {
			c.Writer.Write([]byte(","))
		}
		c.Writer.Write([]byte(toJSON(record)))
		c.Writer.Flush()
		isFirstRecord = false
	}

	// Close the JSON response
	c.Writer.Write([]byte("]}"))
	c.Writer.Flush()
}

func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
