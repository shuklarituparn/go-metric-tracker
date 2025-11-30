package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type loggingresponseWriter struct {
	gin.ResponseWriter
	size int
}

func (w *loggingresponseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, err
}

func (w *loggingresponseWriter) WriteString(s string) (int, error) {
	size, err := w.ResponseWriter.WriteString(s)
	w.size += size
	return size, err
}

// Logger создает middleware для логирования HTTP запросов с использованием zap logger.
// Логирует URI, HTTP метод, статус ответа, время выполнения и размер ответа.
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		path := ctx.Request.URL.Path
		rw := &loggingresponseWriter{ResponseWriter: ctx.Writer}
		ctx.Writer = rw

		ctx.Next()
		latency := time.Since(start)

		logger.Info(
			"",
			zap.String("uri", path),
			zap.String("method", ctx.Request.Method),
			zap.Int("status", ctx.Writer.Status()),
			zap.Duration("duration", latency),
			zap.Int("size", rw.size),
		)
	}
}
