package main

import (
	"cafe_main/internal/auth"
	"cafe_main/internal/logger"
	"strings"

	"os"

	"github.com/gin-gonic/gin"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RefreshToken struct {
	RefreshToken string `json:"refresh_token"`
}

func main() {
	logger := logger.GetLogger()
	secretKey := os.Getenv("SECRET_KEY")

	dbOptions := &auth.DBOptions{
		UserStorage: auth.UserStorageOptions{
			DBPath:        os.Getenv("USER_DB_PATH"),
			AdminLogin:    os.Getenv("ADMIN_LOGIN"),
			AdminPassword: os.Getenv("ADMIN_PASSWORD"),
		},
		SessionStorage: auth.SessionStorageOptions{
			Addr:     os.Getenv("REDIS_ADDR"),
			Password: "",
			DB:       0,
		},
	}

	authServices, err := auth.NewAuthService(dbOptions, logger, secretKey)

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

		c.JSON(200, gin.H{"access_token": tokenString.AccessToken, "refresh_token": tokenString.RefreshToken})
	})

	router.POST("/logout", func(c *gin.Context) {

		c.JSON(200, gin.H{"message": "Logout successful"})
	})

	router.POST("/refresh", func(c *gin.Context) {
		var refreshToken RefreshToken
		err := c.BindJSON(&refreshToken)

		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		tokens, err := authServices.RefreshToken(refreshToken.RefreshToken)

		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid refresh token"})
			return
		}

		c.JSON(200, gin.H{"access_token": tokens.AccessToken, "refresh_token": tokens.RefreshToken})
	})

	AuthMiddleware := func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Missing Authorization header"})
			c.Abort()
			return
		}

		tokenString := strings.Split(authHeader, " ")[1]

		_, err := authServices.CheckToken(tokenString)

		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Next()
	}

	v1 := router.Group("/v1")
	v1.Use(AuthMiddleware)
	v1.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router.Run() // listen and serve on 0.0.0.0:8080
}
