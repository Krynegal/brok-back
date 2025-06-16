package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"brok/internal/utils"
)

// JWTAuth - middleware для аутентификации с использованием JWT
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Читаем заголовок Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header missing"})
			return
		}

		// Разделяем заголовок по пробелу (Bearer <token>)
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		tokenString := parts[1]

		// Используем ParseJWT для парсинга и валидации токена
		claims, err := utils.ParseJWT(tokenString) // Применяем вашу функцию
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// Прокладываем user_id из claims в контекст Gin
		c.Set("user_id", claims.UserID)

		// Переходим к следующему обработчику
		c.Next()
	}
}
