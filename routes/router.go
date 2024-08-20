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
	
	r.POST("/auth/login", userController.Login)
	
	r.POST("/forgot-password", userController.InitiatePasswordReset)
	r.POST("/reset-password", userController.ResetPassword)

	r.GET("/mappings", mappingController.GetMappings)
	r.GET("/mapping/data", mappingController.GetAllMapping)
	r.POST("/mapping/insert", mappingController.InsertMapping)
	r.PUT("/mapping/update/:id", mappingController.UpdateMapping)
	r.POST("/mapping/delete/:id", mappingController.DeleteMapping)

	r.GET("/mapping/:id", mappingController.GetMapping)

	r.GET("/kmz/:id", mappingController.GetKMZFile)


	r.GET("/user/data", userController.GetUsers)
	r.POST("/user/insert", userController.InsertUser)


	r.GET("/vessel/data", vesselController.GetVessel)
	r.GET("/vessel/:call_sign", vesselController.GetVesselByCallSign)
	r.POST("/vessel/insert", vesselController.InsertVessel)
	r.PUT("/vessel/update/:call_sign", vesselController.UpdateVessel)
	r.POST("/vessel/delete/:call_sign", vesselController.DeleteVessel)

	r.GET("/vessel_records/:call_sign", vesselController.GetVesselRecords)
	

	r.POST("/user/sendOtp", otpController.InsertOtp)

}
