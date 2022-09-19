package logging

import (
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func RequestLogging(logger *zap.Logger) gin.HandlerFunc {
	return ginzap.GinzapWithConfig(logger, &ginzap.Config{
		UTC:        true,
		TimeFormat: time.RFC3339,
		Context: func(context *gin.Context) []zapcore.Field {
			var fields []zapcore.Field
			// log request ID
			if requestID := context.Writer.Header().Get("X-Request-Id"); requestID != "" {
				fields = append(fields, zap.String("request_id", requestID))
			}

			// log trace and span ID
			if trace.SpanFromContext(context.Request.Context()).SpanContext().IsValid() {
				fields = append(fields, zap.String("trace_id",
					trace.SpanFromContext(context.Request.Context()).SpanContext().TraceID().String()))
				fields = append(fields, zap.String("span_id",
					trace.SpanFromContext(context.Request.Context()).SpanContext().SpanID().String()))
			}

			return fields
		},
	})
}
