package controllers

import (
	"fmt"
	"golang-app/app/models"
	"golang-app/database"
	"net/http"

	"github.com/gin-gonic/gin"

)

type PendudukController struct {
	// Dependent services
}

func NewPendudukController() *PendudukController {
	return &PendudukController{
		// Inject services
	}
}
func (r *PendudukController) Index(c *gin.Context) {
	// Data to pass to the index.html template

	var users []models.User

	// Retrieve all users from the database
	err := database.DB.Find(&users).Error

	if err != nil {
		fmt.Println(err)
	}

	app := gin.H{
		"title": "Dashboard Panel",
		"users": users,
	}

	c.HTML(http.StatusOK, "penduduk.html", app)
}
