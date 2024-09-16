package controllers

import (
	"fmt"
	"golang-app/app/models"
	"golang-app/database"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

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
	if err := database.DB.Find(&mappings).Error; err != nil {
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
	result := database.DB.Find(&data)
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

func (r *MappingController) GetMapping(c *gin.Context) {
	id := c.Param("id")
	var mapping models.Mapping
	if err := database.DB.First(&mapping, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Overlay not found"})
		return
	}
	c.JSON(http.StatusOK, mapping)
}

func (r *MappingController) InsertMapping(c *gin.Context) {
	var input struct {
		Name   string `form:"name" binding:"required"`
		Status bool   `form:"status" binding:"required"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to bind data", "error": err.Error()})
		return
	}

	// Handle file upload
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to get file", "error": err.Error()})
		return
	}

	// Validate file extension
	if !isValidFileType(file.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid file type. Only KML and KMZ files are allowed."})
		return
	}

	// Generate filename using datetime and input name
	filename := generateFilename(input.Name, filepath.Ext(file.Filename))

	// Save the file
	if err := c.SaveUploadedFile(file, filepath.Join("public/upload/assets/image/mapping", filename)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to save file", "error": err.Error()})
		return
	}

	// Create mapping struct
	mapping := models.Mapping{
		Name:   input.Name,
		File:   filename,
		Status: input.Status,
	}

	// Insert the new mapping into the database
	err = database.DB.Create(&mapping).Error
	if err != nil {
		// Handle potential errors
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create Overlay", "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Overlay created successfully", "data": mapping})
}

func (r *MappingController) UpdateMapping(c *gin.Context) {
	// Get the mapping ID from the URL parameter
	id := c.Param("id")

	// Find the existing mapping
	var mapping models.Mapping
	if err := database.DB.First(&mapping, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Mapping not found"})
		return
	}

	// Bind the input data
	var input struct {
		Name   string `form:"name" binding:"required"`
		Status bool   `form:"status"`
	}
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "error": err.Error()})
		return
	}

	// Update fields
	mapping.Name = input.Name
	mapping.Status = input.Status

	// Handle file upload if a new file is provided
	file, err := c.FormFile("file")
	if err == nil {
		// A new file was uploaded
		// Validate file extension
		ext := strings.ToLower(path.Ext(file.Filename))
		if ext != ".kml" && ext != ".kmz" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid file type. Only KML and KMZ files are allowed."})
			return
		}

		// Generate a unique filename
		filename := generateFilename(input.Name, path.Ext(file.Filename))
		filePath := path.Join("public/upload/assets/image/mapping", filename)

		// Save the file
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to save file", "error": err.Error()})
			return
		}

		// Delete the old file if it exists
		if mapping.File != "" {
			oldFilePath := path.Join("public/upload/assets/image/mapping", mapping.File)
			os.Remove(oldFilePath) // Ignore errors, as the file might not exist
		}

		mapping.File = filename
	}

	// Save the updated mapping to the database
	if err := database.DB.Save(&mapping).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update Overlay", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Overlay updated successfully", "data": mapping})
}

func (r *MappingController) DeleteMapping(c *gin.Context) {
	// Get the mapping ID from the URL parameter
	id := c.Param("id")

	var input struct {
		ConfirmationName string `form:"confirmationName" json:"confirmationName" binding:"required"`
	}
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "error": err.Error()})
		return
	}

	// Get the confirmation name from the form data
	log.Printf("Received form data: %+v", input.ConfirmationName)
	confirmationName := input.ConfirmationName
	if confirmationName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Confirmation name is required"})
		return
	}

	// Find the mapping in the database
	var mapping models.Mapping
	if err := database.DB.First(&mapping, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Overlay not found"})
		return
	}

	// Check if the confirmation name matches the actual mapping name
	if confirmationName != mapping.Name {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Confirmation name does not match the Overlay name"})
		return
	}

	// Get the file path
	filePath := filepath.Join("public/upload/assets/image/mapping", mapping.File)

	// Delete the file from the server
	if err := os.Remove(filePath); err != nil {
		// If the file doesn't exist, log it but continue with the database deletion
		if !os.IsNotExist(err) {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete file", "error": err.Error()})
			return
		}
		// Log that the file was not found
		// You might want to use a proper logging library here
		log.Printf("File not found: %s", filePath)
	}

	// Delete the mapping from the database
	if err := database.DB.Delete(&mapping).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete Overlay from database", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Overlay deleted successfully"})
}

// Helper function to generate a filename using datetime and input name
func generateFilename(name string, ext string) string {
	// Remove any spaces from the name and replace with underscores
	sanitizedName := strings.ReplaceAll(name, " ", "_")
	// Get current date and time
	now := time.Now()
	// Format: YYYYMMDD_HHMMSS_name.ext
	return fmt.Sprintf("%s_%s%s", now.Format("20060102_150405"), sanitizedName, ext)
}

// Helper function to validate file type
func isValidFileType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".kml" || ext == ".kmz"
}
