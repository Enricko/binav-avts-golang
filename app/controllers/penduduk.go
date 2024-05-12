package controllers

import (
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
	app := gin.H{
		"title": "Dashboard Panel",
	}

	c.HTML(http.StatusOK, "penduduk.html", app)
}
