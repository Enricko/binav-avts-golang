package controllers

import (
	"golang-app/app/models"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
)

const (
	chunkSize = 4 * 1024 * 1024                        // 4MB chunks
	kmzDir    = "./public/upload/assets/image/mapping" // Directory containing KMZ files
)

type Controller struct {
	// Dependent services
}

func NewController() *Controller {
	return &Controller{}
}

func (r *Controller) Index(c *gin.Context) {
	isLoggedInInterface, exists := c.Get("isLoggedIn")
	isLoggedIn := false
	if exists {
		isLoggedIn, _ = isLoggedInInterface.(bool)
	}

	var user models.User
	var permissions map[string]bool

	if isLoggedIn {
		userInterface, exists := c.Get("user")
		if exists {
			user, _ = userInterface.(models.User)
		}

		permissions = map[string]bool{
			"canViewBasicInfo":   true,
			"canUseAdminTools":   user.Level == models.ADMIN || user.Level == models.OWNER,
			"canConfigureSystem": user.Level == models.OWNER,
		}
	} else {
		permissions = map[string]bool{
			"canViewBasicInfo":   false,
			"canUseAdminTools":   false,
			"canConfigureSystem": false,
		}
	}

	data := gin.H{
		"title":       "Binav AVTS",
		"csrfToken":   csrf.GetToken(c),
		"isLoggedIn":  isLoggedIn,
		"user":        user,
		"permissions": permissions,
	}

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

func (r *Controller) KmzInfo(c *gin.Context) {
	filename := c.Query("file")
	if filename == "" {
		log.Printf("KmzInfo: No file specified")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file specified"})
		return
	}

	filePath := filepath.Join(kmzDir, filename)
	log.Printf("KmzInfo: Attempting to open file: %s", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("KmzInfo: Failed to open KMZ file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open KMZ file"})
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		log.Printf("KmzInfo: Failed to get file info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file info"})
		return
	}

	totalChunks := (fileInfo.Size() + chunkSize - 1) / chunkSize
	log.Printf("KmzInfo: File size: %d bytes, Total chunks: %d", fileInfo.Size(), totalChunks)

	c.JSON(http.StatusOK, gin.H{
		"totalChunks": totalChunks,
		"chunkSize":   chunkSize,
	})
}

// New method for KMZ chunk
func (r *Controller) KmzChunk(c *gin.Context) {
	filename := c.Query("file")
	if filename == "" {
		log.Printf("KmzChunk: No file specified")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file specified"})
		return
	}

	chunkStr := c.Param("chunk")
	chunk, err := strconv.ParseInt(chunkStr, 10, 64)
	if err != nil {
		log.Printf("KmzChunk: Invalid chunk number: %s", chunkStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chunk number"})
		return
	}

	filePath := filepath.Join(kmzDir, filename)
	log.Printf("KmzChunk: Attempting to open file: %s", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("KmzChunk: Failed to open KMZ file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open KMZ file"})
		return
	}
	defer file.Close()

	_, err = file.Seek(chunk*chunkSize, 0)
	if err != nil {
		log.Printf("KmzChunk: Failed to seek in file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to seek in file"})
		return
	}

	buffer := make([]byte, chunkSize)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		log.Printf("KmzChunk: Failed to read file chunk: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file chunk"})
		return
	}

	log.Printf("KmzChunk: Successfully read chunk %d, size: %d bytes", chunk, n)
	c.Data(http.StatusOK, "application/octet-stream", buffer[:n])
}
