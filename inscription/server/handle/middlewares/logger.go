package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/inscription-c/cins/inscription/log"
	"time"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		// Process request
		c.Next()
		if raw != "" {
			path = path + "?" + raw
		}
		log.Srv.Infof("method: %s, path: %s, status: %d, latency: %s, client_ip: %s, error_message: %s, body_size: %d",
			c.Request.Method, path, c.Writer.Status(), time.Now().Sub(start), c.ClientIP(), c.Errors.ByType(gin.ErrorTypePrivate).String(), c.Writer.Size())
	}
}
