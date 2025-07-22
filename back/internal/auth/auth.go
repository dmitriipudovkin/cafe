package auth

import (
	"fmt"

	"cafe_main/internal/auth/storage"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func AuthMiddleware(c *gin.Context) {
	fmt.Println("Im a dummy!")

	c.Next()
}

// TO DO надо понять как это все нормально попилить
func NewAuthStorage(dbPath string, logger *logrus.Logger) (*storage.AuthStorage, error) {
	return storage.NewAuthStorage(dbPath, logger)
}
