package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"time"
)

func Recording(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		traceId := trace.SpanContextFromContext(c.Request.Context()).TraceID().String()
		spanId := trace.SpanContextFromContext(c.Request.Context()).SpanID().String()
		c.Header("Request-Id", traceId)
		// 设置 error 默认值为 ok
		c.Set("error", "ok")
		c.Next()
		
		err, _ := c.Get("error")
		
		logger.Info(err.(string),
			zap.String("traceId", traceId),
			zap.String("spanId", spanId),
			zap.Int64("startTimestamp", start.Unix()),
			zap.String("token", c.GetHeader("token")),
			zap.String("clientIP", c.ClientIP()),
			zap.String("requestURI", c.Request.RequestURI),
			zap.String("contentType", c.ContentType()),
			zap.String("method", c.Request.Method),
			zap.String("host", c.Request.Host),
			zap.String("form", c.Request.Form.Encode()),
			//zap.String("traceId", c.Request.Header.),
			zap.Float64("latency", time.Now().Sub(start).Seconds()),
			zap.Int("status", c.Writer.Status()),
			zap.Int("size", c.Writer.Size()),
		)
	}
}
