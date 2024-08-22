package middleware

import (
	"fmt"
	"golang-app/app/models"
	"golang-app/database"
	"net/http"
	"net/url"
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

func UserAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("token")
		if err != nil {
			c.Set("isLoggedIn", false)
			c.Next()
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil // Replace with your actual secret key
		})

		if err != nil || !token.Valid {
			c.Set("isLoggedIn", false)
			c.Next()
			return
		}

		var user models.User
		result := database.DB.Where("email = ?", claims.Email).First(&user)
		if result.Error != nil {
			c.Set("isLoggedIn", false)
			c.Next()
			return
		}

		c.Set("isLoggedIn", true)
		c.Set("user", user)
		c.Next()
	}
}
func AlreadyLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("token")
		if err == nil {
			// Token found, let's validate it
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return jwtKey, nil
			})

			if err == nil && token.Valid {
				// User is already logged in, redirect to home page
				redirectToAlert(c, "/", "info", "You are already logged in")
				return
			}
		}

		// If no valid token is found, the user is not logged in
		// Allow them to proceed to the login page
		c.Next()
	}
}

func IsLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("token")
		if err != nil {
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				redirectToAlert(c, "/login", "unauthorized", "Not logged in")
				return
			}
			splitToken := strings.Split(authHeader, "Bearer ")
			if len(splitToken) != 2 {
				redirectToAlert(c, "/login", "unauthorized", "Invalid token format")
				return
			}
			tokenString = splitToken[1]
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			redirectToAlert(c, "/login", "unauthorized", "Invalid or expired token")
			return
		}

		c.Next()
	}
}

func CheckUserLevel(requiredLevel models.Level) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("token")
		if err != nil {
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				redirectToAlert(c, "/login", "unauthorized", "Authorization token required")
				return
			}	
			splitToken := strings.Split(authHeader, "Bearer ")
			if len(splitToken) != 2 {
				redirectToAlert(c, "/login", "unauthorized", "Invalid token format")
				return
			}
			tokenString = splitToken[1]
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			redirectToAlert(c, "/login", "unauthorized", "Invalid or expired token")
			return
		}

		if !isLevelSufficient(claims.Level, requiredLevel) {
			redirectToPreviousWithAlert(c, "insufficient", "Insufficient privileges")
			return
		}

		c.Set("email", claims.Email)
		c.Set("level", claims.Level)
		c.Next()
	}
}

func redirectToAlert(c *gin.Context, path, alertType, message string) {
	c.Redirect(http.StatusFound, fmt.Sprintf("%s?alert=%s&message=%s", path, alertType, message))
	c.Abort()
}

func redirectToPreviousWithAlert(c *gin.Context, alertType, message string) {
	referer := c.Request.Header.Get("Referer")
	if referer == "" {
		referer = "/"
	}

	u, err := url.Parse(referer)
	if err != nil {
		u, _ = url.Parse("/")
	}

	q := u.Query()
	q.Set("alert", alertType)
	q.Set("message", message)
	u.RawQuery = q.Encode()

	c.Redirect(http.StatusFound, u.String())
	c.Abort()
}

func isLevelSufficient(userLevel, requiredLevel models.Level) bool {
	levelHierarchy := map[models.Level]int{
		models.USER:  1,
		models.ADMIN: 2,
		models.OWNER: 3,
	}

	return levelHierarchy[userLevel] >= levelHierarchy[requiredLevel]
}
