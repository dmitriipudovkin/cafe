package main

import (
	"cafe_main/internal/auth"
	"cafe_main/internal/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	logger := logger.GetLogger()

	auth.NewAuthStorage("./auth.db", logger)

	router := gin.Default()

	router.POST("/login", func(c *gin.Context) {

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
