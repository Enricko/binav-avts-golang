package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"golang-app/app/controllers"
	"golang-app/app/middleware"
	"golang-app/database"
	"golang-app/routes"
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
	// Define a flag for setting the mode
	mode := flag.String("mode", "debug", "Set the runtime mode (debug/release)")
	flag.Parse()

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using existing environment variables")
	}

	// Check for MODE environment variable, override flag if present
	if envMode := os.Getenv("GIN_MODE"); envMode != "" {
		*mode = envMode
	}

	// Set Gin mode
	switch *mode {
	case "debug", "development":
		gin.SetMode(gin.DebugMode)
		log.Println("Running in Debug/Development mode")
	case "release", "production":
		gin.SetMode(gin.ReleaseMode)
		log.Println("Running in Release/Production mode")
	default:
		log.Printf("Unknown mode: %s, defaulting to Debug mode", *mode)
		gin.SetMode(gin.DebugMode)
	}

	// Initialize database
	database.Init()
	
	defer database.DB.Close()

	// Initialize router
	r := gin.New()

	// Use Logger and Recovery middleware
	if gin.Mode() == gin.DebugMode {
		r.Use(gin.Logger())
	}
	r.Use(gin.Recovery())

	// CORS configuration
	corsConfig := cors.DefaultConfig()
	if gin.Mode() == gin.ReleaseMode {
		corsConfig.AllowOrigins = []string{"https://binav-avts.id", "https://www.binav-avts.id"}
	} else {
		corsConfig.AllowAllOrigins = true
	}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	corsConfig.ExposeHeaders = []string{"Content-Length"}
	corsConfig.AllowCredentials = true
	corsConfig.MaxAge = 12 * time.Hour
	r.Use(cors.New(corsConfig))

	// Session configuration
	store := cookie.NewStore([]byte(os.Getenv("SECRET_KEY")))
	store.Options(sessions.Options{
		HttpOnly: true,
		Secure:   gin.Mode() == gin.ReleaseMode, // Secure in production
		SameSite: http.SameSiteStrictMode,
	})
	r.Use(sessions.Sessions("mysession", store))

	// CSRF protection
	excludedPaths := []string{
		"/vessel_ip",
		// Add paths to exclude from CSRF protection
	}
	r.Use(middleware.NoCSRF(excludedPaths))

	// Initialize controllers
	telnetController := controllers.NewTelnetController()
	webSocketController := controllers.NewWebSocketController(telnetController)

	// Start Telnet connections
	go func() {
		telnetController.StartTelnetConnections()
	}()

	// Set up routes
	r.GET("/ws", webSocketController.HandleWebSocket)
	r.LoadHTMLGlob("templates/**/*")
	routes.SetupRouter(r)
	r.Static("/public", "./public")

	// Conditional debugging tools
	if gin.Mode() == gin.DebugMode {
		// Add development-specific routes or middleware here
		r.GET("/debug/vars" /* ... */)
	}

	// Server configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:    "0.0.0.0:" + port,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
