package main

import (
	"golang-app/routes"
	"golang-app/database"
    "log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

)

func main() {
	err := godotenv.Load(".env")
    if err != nil {
        log.Fatal("Error loading .env file")
    }

	database.Init()

	r := gin.Default()

	r.LoadHTMLGlob("templates/**/*")
	routes.SetupRouter(r)

	r.Static("/public", "./public")
	
	r.Run("127.0.0.1:8080")
	
	defer database.DB.Close()
}