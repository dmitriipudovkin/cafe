package main

import (
	"cafe_main/internal/auth"
	"cafe_main/internal/logger"

	"os"

	"github.com/gin-gonic/gin"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	logger := logger.GetLogger()
	secretKey := os.Getenv("SECRET_KEY")

	authServices, err := auth.NewAuthService("./auth.db", logger, secretKey)

	if err != nil {
		logger.Fatal(err)
		os.Exit(1)
	}

	router := gin.Default()

	router.POST("/login", func(c *gin.Context) {
		var user User
		err := c.BindJSON(&user)

		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		tokenString, err := authServices.Login(user.Username, user.Password)

		if err != nil {
			c.JSON(401, gin.H{"error": "Wrong credentials"})
			return
		}

		c.JSON(200, gin.H{"token": tokenString})
	})

	v1 := router.Group("/v1")
	v1.Use(auth.AuthMiddleware)
	v1.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router.Run() // listen and serve on 0.0.0.0:8080
}
