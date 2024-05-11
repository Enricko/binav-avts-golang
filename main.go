package main

import (
	"golang-app/routes"

	"github.com/gin-gonic/gin"

)

func main() {
	r := gin.Default()

	r.LoadHTMLGlob("templates/**/*")
	routes.SetupRouter(r)

	r.Run("127.0.0.1:8080")
}