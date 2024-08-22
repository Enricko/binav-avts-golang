package controllers

import (
	"golang-app/app/models"
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
