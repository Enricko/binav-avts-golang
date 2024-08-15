package routes

import (
	"golang-app/app/controllers"

	"github.com/gin-gonic/gin"

)

func SetupRouter(r *gin.Engine) {

	mainController := controllers.NewController()
	userController := controllers.NewUserController()
	mappingController := controllers.NewMappingController()
	vesselController := controllers.NewVesselController()
	otpController := controllers.NewOtpController()

	r.GET("/", mainController.Index)
	r.GET("/login", mainController.Login)
	
	r.POST("/login", userController.Login)

	r.GET("/mappings", mappingController.GetMappings)
	r.GET("/mapping/data", mappingController.GetAllMapping)
	r.POST("/mapping/data/submit", mappingController.InsertMapping)
	r.GET("/kmz/:id", mappingController.GetKMZFile)


	r.GET("/user/data", userController.GetUsers)
	r.POST("/user/insert", userController.InsertUser)


	r.GET("/vessel/data", vesselController.GetVessel)
	r.GET("/vessel/:call_sign", vesselController.GetVesselByCallSign)
	r.POST("/vessel/insert", vesselController.InsertVessel)
	r.PUT("/vessel/update/:call_sign", vesselController.UpdateVessel)
	r.DELETE("/vessel/delete/:call_sign", vesselController.DeleteVessel)

	r.GET("/vessel_records/:call_sign", vesselController.GetVesselRecords)
	

	r.POST("/user/sendOtp", otpController.InsertOtp)

}
