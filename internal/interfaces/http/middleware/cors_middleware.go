package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS returns a CORS middleware backed by an allow-list of origins.
//
// When isProduction is false and origins is empty, it falls back to the
// permissive "*" behaviour for local development. In production an empty
// allow-list is enforced upstream by config validation, so this should never
// happen — but to fail closed we still refuse the wildcard here.
func CORS(origins []string, isProduction bool) gin.HandlerFunc {
	allowAll := false
	allowed := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		o = strings.TrimSpace(o)
		if o == "" {
			continue
		}
		if o == "*" && !isProduction {
			allowAll = true
			continue
		}
		allowed[o] = struct{}{}
	}
	if !isProduction && len(allowed) == 0 && !allowAll {
		allowAll = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		h := c.Writer.Header()

		if origin != "" {
			if allowAll {
				h.Set("Access-Control-Allow-Origin", "*")
			} else if _, ok := allowed[origin]; ok {
				h.Set("Access-Control-Allow-Origin", origin)
				h.Set("Vary", "Origin")
				h.Set("Access-Control-Allow-Credentials", "true")
			}
		}

		h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		h.Set("Access-Control-Max-Age", "600")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
