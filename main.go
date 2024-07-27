package main

import (
	"golang-app/app/controllers"
	"golang-app/database"
	"golang-app/routes"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	csrf "github.com/utrack/gin-csrf"
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
	r.Use(CORSMiddleware())

	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	r.Use(csrf.Middleware(csrf.Options{
		Secret: os.Getenv("SECRET"),
		ErrorFunc: func(c *gin.Context) {
			c.String(http.StatusBadRequest, "CSRF token mismatch")
			c.Abort()
		},
	}))

	telnetController := controllers.NewTelnetController()

	// Start Telnet connections in a separate goroutine
	go telnetController.StartTelnetConnections()

	// telnetController := controllers.NewTelnetController()

	r.GET("/ws/kapal", telnetController.KapalTelnetWebsocketHandler)

	r.LoadHTMLGlob("templates/**/*")
	routes.SetupRouter(r)

	r.Static("/public", "./public")

	r.Run("0.0.0.0:8080")

	defer database.DB.Close()
}
