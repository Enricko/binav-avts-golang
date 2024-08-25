package main

import (
	"golang-app/app/controllers"
	"golang-app/app/middleware"
	"golang-app/database"
	"golang-app/routes"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	database.Init()

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	// r.Use(CORSMiddleware())

	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	excludedPaths := []string{
		// Add more paths as needed
	} 

	// Use the NoCSRF middleware with the exclusion list
	r.Use(middleware.NoCSRF(excludedPaths))

	telnetController := controllers.NewTelnetController()

	// Start Telnet connections in a separate goroutine
	go telnetController.StartTelnetConnections()

	// r.GET("/ws/kapal", telnetController.KapalTelnetWebsocketHandler)
	webSocketController := controllers.NewWebSocketController(telnetController)

	// Set up the WebSocket route
	r.GET("/ws", webSocketController.HandleWebSocket)

	r.LoadHTMLGlob("templates/**/*")
	routes.SetupRouter(r)

	r.Static("/public", "./public")

	r.Run("0.0.0.0:8080")

	defer database.DB.Close()
}