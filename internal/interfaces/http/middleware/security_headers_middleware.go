package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders sets a baseline of HTTP response headers that are safe for
// a JSON API. HSTS is only emitted when the request was served over TLS (or
// behind a TLS-terminating proxy that sets X-Forwarded-Proto=https) so that
// local development over plain HTTP still works.
func SecurityHeaders(isProduction bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		h.Set("Cross-Origin-Resource-Policy", "same-site")
		h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		if isProduction && isHTTPS(c) {
			h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		}

		c.Next()
	}
}

func isHTTPS(c *gin.Context) bool {
	if c.Request.TLS != nil {
		return true
	}
	if proto := c.GetHeader("X-Forwarded-Proto"); proto == "https" {
		return true
	}
	return false
}
