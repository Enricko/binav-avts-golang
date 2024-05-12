package routes

import (
	"golang-app/app/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) {

	dashboardController := controllers.NewDashboardController()

	r.GET("/dashboard", dashboardController.Index)

	r.GET("/user", dashboardController.User)

	pendudukController := controllers.NewPendudukController()
	r.GET("/penduduk", pendudukController.Index)

	r.GET("/penduduk/data", pendudukController.DataPenduduk)
}
