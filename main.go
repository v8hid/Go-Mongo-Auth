package main

import (
	"log"
	"os"

	routes "v8hid/Go-Mongo-Auth/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	router := gin.New()
	router.SetTrustedProxies(nil)
	router.Use(gin.Logger())

	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	router.GET("/api-test", func(ctx *gin.Context) {
		email, _ := ctx.Get("email")
		role, _ := ctx.Get("role")

		ctx.JSON(200, gin.H{
			"status": "success",
			"email":  email,
			"role":   role,
		})

	})

	router.Run(":" + port)
}
