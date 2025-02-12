package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"todo-app/config"
	"todo-app/controllers"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware untuk validasi JWT token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ambil token dari header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		// Pastikan token menggunakan format "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		tokenString := parts[1] // Ambil tokennya saja

		// Parse token dengan claims
		claims := &controllers.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.JwtKey), nil
		})

		if err != nil || !token.Valid {
			fmt.Println("JWT Parsing Error:", err) // Log error jika parsing gagal
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "details": err.Error()})
			c.Abort()
			return
		}

		// Simpan UserID dalam context untuk digunakan di handler lain
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
