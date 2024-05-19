package main

import (
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

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	database.Init()

	r := gin.Default()

	store := cookie.NewStore([]byte(os.Getenv("SECRET")))
	r.Use(sessions.Sessions("mysession", store))

	r.Use(csrf.Middleware(csrf.Options{
		Secret: os.Getenv("SECRET"),
		ErrorFunc: func(c *gin.Context) {
			c.String(http.StatusBadRequest, "CSRF token mismatch")
			c.Abort()
		},
	}))

	r.LoadHTMLGlob("templates/**/*")
	routes.SetupRouter(r)

	r.Static("/public", "./public")

	r.Run("127.0.0.1:8080")

	defer database.DB.Close()
}
