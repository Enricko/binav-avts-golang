package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
)

// NoCSRF excludes specific paths from CSRF protection
func NoCSRF(excludedPaths []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, path := range excludedPaths {
			if strings.HasPrefix(c.Request.URL.Path, path) {
				// Skip CSRF protection for this route
				c.Next()
				return
			}
		}

		// Apply CSRF protection for all other routes
		csrf.Middleware(csrf.Options{
			Secret: os.Getenv("SECRET"),
			ErrorFunc: func(c *gin.Context) {
				c.String(http.StatusBadRequest, "CSRF token mismatch")
				c.Abort()
			},
		})(c)
	}
}
