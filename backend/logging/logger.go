package logging

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

const requestIDKey = "request_id"

type Fields map[string]interface{}

func Info(event, message string, fields Fields) {
	write("INFO", event, message, fields)
}

func Warn(event, message string, fields Fields) {
	write("WARN", event, message, fields)
}

func Error(event, message string, fields Fields) {
	write("ERROR", event, message, fields)
}

func RequestIDFromGin(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if v, ok := c.Get(requestIDKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = newRequestID()
		}

		c.Set(requestIDKey, requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		start := time.Now()
		Info("request.start", "incoming request", Fields{
			"request_id": requestID,
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"client_ip":  c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		})

		c.Next()

		durationMs := time.Since(start).Milliseconds()
		fields := Fields{
			"request_id":    requestID,
			"method":        c.Request.Method,
			"path":          c.Request.URL.Path,
			"status":        c.Writer.Status(),
			"duration_ms":   durationMs,
			"response_size": c.Writer.Size(),
			"error_count":   len(c.Errors),
		}

		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
			Error("request.end", "request finished with errors", fields)
			return
		}

		Info("request.end", "request finished", fields)
	}
}

func write(level, event, message string, fields Fields) {
	entry := map[string]interface{}{
		"ts":      time.Now().Format(time.RFC3339Nano),
		"level":   level,
		"event":   event,
		"message": message,
	}
	for k, v := range fields {
		entry[k] = v
	}

	payload, err := json.Marshal(entry)
	if err != nil {
		log.Printf("{\"ts\":%q,\"level\":%q,\"event\":%q,\"message\":%q,\"marshal_error\":%q}", time.Now().Format(time.RFC3339Nano), level, event, message, err.Error())
		return
	}
	log.Printf("%s", payload)
}

func newRequestID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("req_%d_%s", time.Now().UnixNano(), hex.EncodeToString(buf))
}
