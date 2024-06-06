package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"

)

type Controller struct {
	// Dependent services
}

func NewController() *Controller {
	return &Controller{}
}

func (r *Controller) Index(c *gin.Context) {
	// Data to pass to the index.html template
	data := gin.H{
		"title":     "Welcome Administrator",
		"csrfToken": csrf.GetToken(c),
	}
	// Render the index.html template with data
	c.HTML(http.StatusOK, "index.html", data)
}
