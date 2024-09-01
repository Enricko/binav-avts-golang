package controllers

import (
	"crypto/rand"
	"fmt"
	"golang-app/app/helper"
	"golang-app/app/models"
	"golang-app/database"
	"math/big"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
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
	tokenBlacklist     map[string]time.Time
	tokenBlacklistLock sync.RWMutex
}

func NewUserController() *UserController {
	controller := &UserController{
		tokenBlacklist: make(map[string]time.Time),
	}
	go controller.cleanupBlacklist()
	return controller
}

type Claims struct {
	Email string       `json:"email"`
	Level models.Level `json:"level"`
	jwt.StandardClaims
}

var jwtKey = []byte(os.Getenv("SECRET_KEY"))

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

func (r *UserController) InsertUser(c *gin.Context) {
	var input struct {
		Name     string       `form:"name" json:"name" binding:"required"`
		Email    string       `form:"email" json:"email" binding:"required,email"`
		Password string       `form:"password" json:"password" binding:"required,min=6"`
		Level    models.Level `form:"level" json:"level" binding:"required"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to bind data", "error": err.Error()})
		return
	}

	// Generate a random ID
	randomID, err := generateRandomString(20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate user ID"})
		return
	}

	// Hash the password
	hashedPassword, err := hashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to hash password"})
		return
	}

	// Validate user level
	if !isValidUserLevel(input.Level) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user level"})
		return
	}

	// Create user struct
	user := models.User{
		IdUser:   randomID,
		Name:     input.Name,
		Email:    input.Email,
		Password: hashedPassword,
		Level:    input.Level,
	}

	// Insert the new user into the database
	err = database.DB.Create(&user).Error
	if err != nil {
		// Check for duplicate email error
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "Duplicate entry") {
			c.JSON(http.StatusConflict, gin.H{"message": "Email already exists"})
			return
		}

		// Handle other potential errors
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create user", "error": err.Error()})
		return
	}
	if err := helper.SendEmail(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to send email"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "data": user})
}

func (r *UserController) SendEmailConfirmation(c *gin.Context) {
	var input struct {
		Name  string `form:"name" json:"name" binding:"required"`
		Email string `form:"email" json:"email" binding:"required,email"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to bind data", "error": err.Error()})
		return
	}

	user := models.User{
		Name:  input.Name,
		Email: input.Email,
	}
	fmt.Println(input)
	fmt.Println(user)

	if err := helper.SendEmail(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to send email","error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Email Sending successfully", "data": user})
}

func (r *UserController) Login(c *gin.Context) {
	var input struct {
		Email    string `form:"email" json:"email" binding:"required,email"`
		Password string `form:"password" json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid email or password"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error finding user"})
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid email or password"})
		return
	}

	expirationHours, err := strconv.Atoi(os.Getenv("JWT_EXPIRATION_HOURS"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid token expiration configuration"})
		return
	}

	expirationTime := time.Now().Add(time.Duration(expirationHours) * time.Hour)
	claims := &Claims{
		Email: input.Email,
		Level: user.Level,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error generating token"})
		return
	}

	// Set token in cookie
	c.SetCookie(
		"token",
		tokenString,
		expirationHours*3600, // convert hours to seconds
		"/",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   tokenString,
		"user":    user,
	})
}

func (r *UserController) InitiatePasswordReset(c *gin.Context) {
	var input struct {
		Email string `form:"email" json:"email" binding:"required,email"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	// Generate OTP
	otp, err := helper.GenerateOTP(6)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate OTP"})
		return
	}

	// Save OTP to the user record
	user.ResetOTP = otp
	user.ResetOTPExpiry = time.Now().Add(15 * time.Minute) // OTP valid for 15 minutes
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to save OTP"})
		return
	}

	// Send OTP to user's email
	if err := helper.SendOTPEmail(user, otp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to send OTP email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully to your email"})
}

func (r *UserController) ValidateOTP(c *gin.Context) {
	var input struct {
		Email string `form:"email" json:"email" binding:"required,email"`
		OTP   string `form:"otp" json:"otp" binding:"required"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	// Check if OTP matches
	if user.ResetOTP != input.OTP {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid OTP"})
		return
	}

	// Check if OTP has expired
	if user.ResetOTPExpiry.IsZero() || time.Now().After(user.ResetOTPExpiry) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "OTP has expired"})
		return
	}

	// OTP is valid
	c.JSON(http.StatusOK, gin.H{"message": "OTP is valid"})
}

func (r *UserController) ResetPassword(c *gin.Context) {
	var input struct {
		Email                string `form:"email" json:"email" binding:"required,email"`
		OTP                  string `form:"otp" json:"otp" binding:"required"`
		NewPassword          string `form:"new_password" json:"new_password" binding:"required,min=6"`
		ConfirmationPassword string `form:"confirmation_password" json:"confirmation_password" binding:"required"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "error": err.Error()})
		return
	}

	// Check if passwords match
	if input.NewPassword != input.ConfirmationPassword {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Passwords do not match"})
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found", "error": err.Error()})
		return
	}

	// Verify OTP
	if user.ResetOTP != input.OTP {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid OTP"})
		return
	}

	// Check if OTP has expired
	if user.ResetOTPExpiry.IsZero() || time.Now().After(user.ResetOTPExpiry) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "OTP has expired"})
		return
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to hash password", "error": err.Error()})
		return
	}

	// Update user's password
	if err := database.DB.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update password", "error": err.Error()})
		return
	}

	// Clear OTP fields
	if err := database.DB.Model(&user).Updates(map[string]interface{}{
		"reset_otp":        "",
		"reset_otp_expiry": time.Now(), // This sets it to the zero value
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to clear OTP fields", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func (r *UserController) Logout(c *gin.Context) {
	// Get token from cookie
	tokenString, err := c.Cookie("token")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No token provided"})
		return
	}

	// Parse the token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}

	// Add the token to the blacklist
	expirationTime := time.Unix(claims.ExpiresAt, 0)
	r.tokenBlacklistLock.Lock()
	r.tokenBlacklist[tokenString] = expirationTime
	r.tokenBlacklistLock.Unlock()

	// Clear the cookie
	c.SetCookie("token", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{"message": "User logged out successfully"})
}

func (r *UserController) cleanupBlacklist() {
	for {
		time.Sleep(1 * time.Hour) // Run cleanup every hour
		now := time.Now()
		r.tokenBlacklistLock.Lock()
		for token, expiry := range r.tokenBlacklist {
			if now.After(expiry) {
				delete(r.tokenBlacklist, token)
			}
		}
		r.tokenBlacklistLock.Unlock()
	}
}

// Helper function to validate user level (unchanged)
func isValidUserLevel(level models.Level) bool {
	return level == models.USER || level == models.ADMIN || level == models.OWNER
}
