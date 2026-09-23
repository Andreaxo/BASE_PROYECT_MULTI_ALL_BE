package middleware

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// sqlInjectionPatterns contains compiled regex patterns for common SQL injection vectors.
// These cover the most dangerous patterns without being overly aggressive for normal inputs.
var sqlInjectionPatterns = []*regexp.Regexp{
	// Classic boolean injection: ' OR '1'='1, ' OR 1=1, etc.
	regexp.MustCompile(`(?i)(\s|'|")(OR|AND)\s+['"]?1['"]?\s*=\s*['"]?1['"]?`),
	// UNION-based injection: UNION SELECT, UNION ALL SELECT
	regexp.MustCompile(`(?i)\bUNION\s+(ALL\s+)?SELECT\b`),
	// Stacked queries / comment-based termination with SQL keywords
	regexp.MustCompile(`(?i)(;|\-\-\s|\/\*).*(SELECT|INSERT|UPDATE|DELETE|DROP|ALTER|CREATE|TRUNCATE|EXEC|EXECUTE|CAST|CONVERT)`),
	// DROP TABLE / DROP DATABASE attacks
	regexp.MustCompile(`(?i)\b(DROP|TRUNCATE|ALTER)\s+(TABLE|DATABASE|SCHEMA|INDEX)\b`),
	// SQL execution functions
	regexp.MustCompile(`(?i)\b(EXEC|EXECUTE|xp_cmdshell|sp_executesql)\b`),
	// Sleep/benchmark timing attacks (blind SQLi)
	regexp.MustCompile(`(?i)\b(SLEEP|BENCHMARK|WAITFOR\s+DELAY|pg_sleep)\s*\(`),
	// Information schema probing
	regexp.MustCompile(`(?i)\binformation_schema\b`),
	// LOAD_FILE / INTO OUTFILE file read/write
	regexp.MustCompile(`(?i)\b(LOAD_FILE|INTO\s+OUTFILE|INTO\s+DUMPFILE)\b`),
}

// containsSQLInjection checks a string for known SQL injection patterns.
func containsSQLInjection(s string) bool {
	for _, pattern := range sqlInjectionPatterns {
		if pattern.MatchString(s) {
			return true
		}
	}
	return false
}

// checkValues recursively scans a JSON-decoded value (string, map, slice) for SQLi patterns.
func checkValues(v interface{}) bool {
	switch val := v.(type) {
	case string:
		return containsSQLInjection(val)
	case map[string]interface{}:
		for _, mv := range val {
			if checkValues(mv) {
				return true
			}
		}
	case []interface{}:
		for _, av := range val {
			if checkValues(av) {
				return true
			}
		}
	}
	return false
}

// SQLInjectionGuard is a Gin middleware that inspects request bodies and query parameters
// for SQL injection patterns and rejects suspicious requests with HTTP 400.
//
// It reads the body non-destructively (by buffering and restoring it) so handlers still
// receive the full body intact.
func SQLInjectionGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ── 1. Check query parameters ────────────────────────────────────────
		for key, values := range c.Request.URL.Query() {
			for _, val := range values {
				if containsSQLInjection(key) || containsSQLInjection(val) {
					log.Printf("[SECURITY] SQLi pattern detected in query param from IP %s: key=%q val=%q", c.ClientIP(), key, val)
					c.JSON(http.StatusBadRequest, gin.H{"error": "Entrada inválida detectada."})
					c.Abort()
					return
				}
			}
		}

		// ── 2. Check URL path parameters ─────────────────────────────────────
		for _, param := range c.Params {
			if containsSQLInjection(param.Value) {
				log.Printf("[SECURITY] SQLi pattern detected in path param from IP %s: key=%q val=%q", c.ClientIP(), param.Key, param.Value)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Entrada inválida detectada."})
				c.Abort()
				return
			}
		}

		// ── 3. Check JSON body ───────────────────────────────────────────────
		contentType := c.GetHeader("Content-Type")
		if strings.HasPrefix(contentType, "application/json") && c.Request.Body != nil && c.Request.ContentLength != 0 {
			// Read the body
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil && len(bodyBytes) > 0 {
				// Restore the body so downstream handlers can still read it
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				// Simple string-level scan on the raw body (fast, catches most cases)
				bodyStr := string(bodyBytes)
				if containsSQLInjection(bodyStr) {
					log.Printf("[SECURITY] SQLi pattern detected in request body from IP %s on %s %s", c.ClientIP(), c.Request.Method, c.Request.URL.Path)
					c.JSON(http.StatusBadRequest, gin.H{"error": "Entrada inválida detectada."})
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}
