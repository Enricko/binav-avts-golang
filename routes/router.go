package routes

import (
	"golang-app/app/controllers"
	"golang-app/app/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) {

	mainController := controllers.NewController()
	userController := controllers.NewUserController()
	mappingController := controllers.NewMappingController()
	vesselController := controllers.NewVesselController()
	otpController := controllers.NewOtpController()

	r.Use(middleware.UserAuthMiddleware())

	r.GET("/login", middleware.AlreadyLoggedIn(), mainController.Login)

	r.POST("/auth/login", middleware.AlreadyLoggedIn(), userController.Login)

	r.POST("/forgot-password", userController.InitiatePasswordReset)
	r.POST("/validate-otp", userController.ValidateOTP)
	r.POST("/reset-password", userController.ResetPassword)

	// telnetController := controllers.NewTelnetController()
	// webSocketController := controllers.NewWebSocketController(telnetController)

	// r.GET("/ws", webSocketController.HandleWebSocket)

	protected := r.Group("/")
	protected.Use(middleware.IsLoggedIn())
	{
		protected.GET("/", mainController.Index)

		protected.POST("/logout", userController.Logout)
		protected.GET("/mappings", mappingController.GetMappings)
		protected.GET("/mapping/data", mappingController.GetAllMapping)
		protected.POST("/mapping/insert", mappingController.InsertMapping)
		protected.PUT("/mapping/update/:id", mappingController.UpdateMapping)
		protected.POST("/mapping/delete/:id", mappingController.DeleteMapping)

		protected.GET("/mapping/:id", mappingController.GetMapping)

		protected.GET("/kmz/:id", mappingController.GetKMZFile)

		protected.GET("/user/data", userController.GetUsers)
		protected.POST("/user/insert", userController.InsertUser)

		protected.GET("/vessel/data", vesselController.GetVessel)
		protected.GET("/vessel/:call_sign", vesselController.GetVesselByCallSign)
		protected.POST("/vessel/insert", vesselController.InsertVessel)
		protected.PUT("/vessel/update/:call_sign", vesselController.UpdateVessel)
		protected.POST("/vessel/delete/:call_sign", vesselController.DeleteVessel)

		protected.GET("/vessel_records/:call_sign", vesselController.GetVesselRecords)
	}

	r.POST("/user/sendOtp", otpController.InsertOtp)

}
