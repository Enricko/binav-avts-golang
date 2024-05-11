package controllers

import (
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

func(r *DashboardController) Index(c *gin.Context){
	// Data to pass to the index.html template
	data := gin.H{
		"title": "Index Page",
	}
	// Render the index.html template with data
	c.HTML(http.StatusOK, "dashboard.html", data)
}

func(r *DashboardController) Home(c *gin.Context){
	// Data to pass to the index.html template
	data := gin.H{
		"title": "Home Page",
	}
	// Render the index.html template with data
	c.HTML(http.StatusOK, "home.html", data)
}