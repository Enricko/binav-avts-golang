package controllers

import (
	"fmt"
	"golang-app/app/models"
	"golang-app/database"
	"net/http"

	"github.com/gin-gonic/gin"

)

type DashboardController struct {
	// Dependent services
}

func NewDashboardController() *DashboardController {
	return &DashboardController{
		// Inject services
	}
}

func (r *DashboardController) Index(c *gin.Context) {
	// Data to pass to the index.html template
	data := gin.H{
		"title": "Index Page",
	}
	// Render the index.html template with data
	c.HTML(http.StatusOK, "dashboard.html", data)
}

func (r *DashboardController) User(c *gin.Context) {
	var users []models.User

	// Retrieve all users from the database
	err := database.DB.Find(&users).Error

	if err != nil {
		fmt.Println(err)
	}

	data := gin.H{
		"title": "Home Page",
		"Users": users,
	}
	// Render the index.html template with data
	c.HTML(http.StatusOK, "user.html", data)
}
