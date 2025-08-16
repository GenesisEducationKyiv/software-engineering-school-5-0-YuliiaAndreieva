package http

import (
	"time"

	"subscription/internal/core/ports/out"

	"github.com/gin-gonic/gin"
)

type MetricsMiddleware struct {
	metrics out.MetricsCollector
}

func NewMetricsMiddleware(metrics out.MetricsCollector) *MetricsMiddleware {
	return &MetricsMiddleware{
		metrics: metrics,
	}
}

func (m *MetricsMiddleware) MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path

		m.metrics.IncrementHTTPRequests(method, path, statusCode)
		m.metrics.RecordHTTPDuration(method, path, duration)
	}
}
