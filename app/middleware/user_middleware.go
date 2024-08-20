package middleware

import (
	"golang-app/app/models"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var jwtKey = []byte(os.Getenv("SECRET_KEY"))

type Claims struct {
	Email string       `json:"email"`
	Level models.Level `json:"level"`
	jwt.StandardClaims
}

func CheckUserLevel(requiredLevel models.Level) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Authorization token required"})
			c.Abort()
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Check if the user's level is sufficient
		if !isLevelSufficient(claims.Level, requiredLevel) {
			c.JSON(http.StatusForbidden, gin.H{"message": "Insufficient privileges"})
			c.Abort()
			return
		}

		c.Set("email", claims.Email)
		c.Set("level", claims.Level)
		c.Next()
	}
}

// Helper function to check if the user's level is sufficient
func isLevelSufficient(userLevel, requiredLevel models.Level) bool {
	levelHierarchy := map[models.Level]int{
		models.USER:  1,
		models.ADMIN: 2,
		models.OWNER: 3,
	}

	return levelHierarchy[userLevel] >= levelHierarchy[requiredLevel]
}
