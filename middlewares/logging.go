package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func LoggingMiddleware() gin.HandlerFunc {
	// logging setup referenced: https://medium.com/pengenpaham/implement-basic-logging-with-gin-and-logrus-5f36fba69b28
	return func(c *gin.Context) {
		// starting time request
		startTime := time.Now()
		// processing request
		c.Next()
		// end time request
		endTime := time.Now()
		// execution time
		latencyTime := endTime.Sub(startTime)
		// request method
		reqMethod := c.Request.Method
		// request route
		reqUri := c.Request.RequestURI
		// status code
		statusCode := c.Writer.Status()
		// request ip
		clientIP := c.ClientIP()

		log.WithFields(log.Fields{
			"METHOD":    reqMethod,
			"URI":       reqUri,
			"STATUS":    statusCode,
			"LATENCY":   latencyTime,
			"CLIENT_IP": clientIP,
		}).Info("HTTP REQUEST")

		c.Next()
	}
}
