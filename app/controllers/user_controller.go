package controllers

import (
	"crypto/rand"
	"fmt"
	"golang-app/app/models"
	"golang-app/database"
	"math/big"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

)

func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// Function to verify a password
func verifyPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func generateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charsetLength := big.NewInt(int64(len(charset)))
	var result strings.Builder

	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return "", err
		}
		result.WriteByte(charset[randomIndex.Int64()])
	}

	return result.String(), nil
}

type UserController struct {
	// Dependent services
}

func NewUserController() *UserController {
	return &UserController{
		// Inject services
	}
}

func (r *UserController) Index(c *gin.Context) {
	// Data to pass to the index.html template
	data := gin.H{
		"title":     "Welcome Administrator",
		"csrfToken": csrf.GetToken(c),
	}
	// Render the index.html template with data
	c.HTML(http.StatusOK, "dashboard.html", data)
}
func (r *UserController) GetUsers(c *gin.Context) {
	// TODO: Change if want to use another model here
	var data []models.User

	// Fetch users data from the database
	database.DB.Find(&data)

	draw, _ := strconv.Atoi(c.Query("draw"))
	start, _ := strconv.Atoi(c.Query("start"))
	length, _ := strconv.Atoi(c.Query("length"))
	search := c.Query("search[value]") // Get search value from DataTables request

	// Filter data based on search query
	// TODO: Change if want to use another model here
	filteredData := make([]models.User, 0)
	for _, user := range data {
		// Iterate over each field of the user struct
		v := reflect.ValueOf(user)
		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i).Interface()
			// Check if the field value contains the search query
			if strings.Contains(strings.ToLower(fmt.Sprintf("%v", fieldValue)), strings.ToLower(search)) {
				// If any field contains the search query, add the user to filtered data and break the loop
				filteredData = append(filteredData, user)
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
				return filteredData[i].IdUser < filteredData[j].IdUser
			})
		} else {
			sort.Slice(filteredData, func(i, j int) bool {
				return filteredData[i].IdUser > filteredData[j].IdUser
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

func (r *UserController) GetUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}

	// Check if user exists
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}
func (r *UserController) GetAllUser(c *gin.Context) {
	// Check if user exists
	var data []models.User
	if err := database.DB.Find(&data).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": data,
	})
}

func (r *UserController) InsertData(c *gin.Context) {
	var requestData map[string]interface{}
	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON"})
		return
	}

	name, nameExists := requestData["name"].(string)
	email, emailExists := requestData["email"].(string)
	password, passwordExists := requestData["password"].(string)
	passwordConfirmation, passwordConfirmationExists := requestData["password_confirmation"].(string)

	// Validate required fields
	if !nameExists || !emailExists || !passwordExists || !passwordConfirmationExists {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Missing required fields"})
		return
	}
	if password != passwordConfirmation {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Password and Password Confirmation must be the same"})
		return
	}

	// Hash the password
	hashedPassword, err := hashPassword(password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Generate id_user (assuming it needs to be unique and not null)
	idUser, err := generateRandomString(25)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Manually create map for the user data
	userData := models.User{
		IdUser:  idUser,
		Name:     name,
		Email:    email,
		Password: hashedPassword,
		// Add other fields as needed
	}

	fmt.Println(userData)

	// Attempt to insert user data into the database
	err = database.DB.Create(&userData).Error
	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "uix_users_email"` {
			c.JSON(http.StatusConflict, gin.H{"message": "Email already exists"})
			return
		}

		// Handle other unique constraint errors or internal server errors
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data Inserted"})

}

func (r *UserController) UpdateData(c *gin.Context) {
	// id := c.Param("id")
	// var input models.User
	// if err := c.ShouldBindJSON(&input); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	// 	return
	// }

	// var user models.User
	// if err := database.DB.First(&user, id).Error; err != nil {
	// 	c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
	// 	return
	// }

	// user.Email = input.Email
	// user.Username = input.Username

	// err := database.DB.Save(&user).Error
	// if err != nil {
	// 	if err.Error() == `pq: duplicate key value violates unique constraint "uix_users_email"` {
	// 		c.JSON(http.StatusConflict, gin.H{"message": "Email already exists"})
	// 		return

	// 	} else if err.Error() == `pq: duplicate key value violates unique constraint "uix_users_username"` {
	// 		c.JSON(http.StatusConflict, gin.H{"message": "Username already exists"})
	// 		return
	// 	}

	// 	c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update user"})
	// 	return
	// }

	// c.JSON(http.StatusOK, gin.H{
	// 	"message": "User updated successfully",
	// })
}
func (r *UserController) DeleteData(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}

	// Check if user exists
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Delete the user
	if err := database.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User Deleted"})
}
