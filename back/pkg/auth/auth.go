package auth

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(c *gin.Context) {
	fmt.Println("Im a dummy!")

	c.Next()
}
