package middleware

import (
	"compress/gzip"
	"io"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

type CompressWriter struct {
	gin.ResponseWriter
	Writer io.Writer
}

func (c *CompressWriter) Write(data []byte) (int, error) {
	return c.Writer.Write(data)
}

func CompressionMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if strings.Contains(ctx.GetHeader("Accept-Encoding"), "gzip") {
			ctx.Header("Content-Encoding", "gzip")
			gz := gzip.NewWriter(ctx.Writer)
			defer func() {
				if err := gz.Close(); err != nil {
					log.Printf("error: while closing the reader: %s", err.Error())
				}
			}()

			ctx.Writer = &CompressWriter{ResponseWriter: ctx.Writer, Writer: gz}
		}
		ctx.Next()
	}
}
