package routes

import (
	"golang-app/app/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) {

	mainController := controllers.NewController()
	r.GET("/", mainController.Index)

	mappingController := controllers.NewMappingController()
	r.GET("/mappings", mappingController.GetMappings)
	r.GET("/mapping/data", mappingController.GetAllMapping)
	r.GET("/kmz/:id", mappingController.GetKMZFile)

	userController := controllers.NewUserController()
	r.GET("/user", userController.Index)

	r.GET("/user/data", userController.GetUsers)
	r.POST("/user/insert", userController.InsertData)

	r.DELETE("/user/delete/:id", userController.DeleteData)
	r.GET("/user/getData/:id", userController.GetUser)
	r.PUT("/user/update/:id", userController.UpdateData)
}
