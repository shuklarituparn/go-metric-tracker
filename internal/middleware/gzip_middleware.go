// Package middleware предоставляет HTTP middleware для обработки сжатия данных.
package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

// GzipWriter оборачивает gin.ResponseWriter для записи сжатых данных.
type GzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

// Write записывает данные в gzip writer.
func (g *GzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

// WriteString записывает строку в gzip writer.
func (g *GzipWriter) WriteString(s string) (int, error) {
	return g.writer.Write([]byte(s))
}

// DecompressionMiddleware создает middleware для декомпрессии входящих запросов.
// Автоматически декомпрессирует тело запроса, если присутствует заголовок Content-Encoding: gzip.
func DecompressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Content-Encoding") == "gzip" {
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.AbortWithStatusJSON(400, gin.H{"error": "Failed to read request body"})
				return
			}
			if err = c.Request.Body.Close(); err != nil {
				log.Printf("failed to close request body: %s", err.Error())
			}

			reader, err := gzip.NewReader(bytes.NewReader(body))
			if err != nil {
				c.AbortWithStatusJSON(400, gin.H{"error": "Failed to decompress request body"})
				return
			}
			defer func() {
				if closeErr := reader.Close(); closeErr != nil {
					log.Printf("error closing reader: %s", closeErr.Error())
				}
			}()

			decompressed, err := io.ReadAll(reader)
			if err != nil {
				c.AbortWithStatusJSON(400, gin.H{"error": "Failed to read decompressed data"})
				return
			}

			c.Request.Body = io.NopCloser(bytes.NewReader(decompressed))
			c.Request.ContentLength = int64(len(decompressed))
			c.Request.Header.Del("Content-Encoding")
		}

		c.Next()
	}
}

// CompressionMiddleware создает middleware для сжатия исходящих ответов.
// Автоматически сжимает ответы с помощью gzip, если клиент поддерживает это (Accept-Encoding: gzip).
func CompressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		acceptEncoding := c.GetHeader("Accept-Encoding")
		if !strings.Contains(acceptEncoding, "gzip") {
			c.Next()
			return
		}

		c.Writer.Header().Set("Content-Encoding", "gzip")
		gz, err := gzip.NewWriterLevel(c.Writer, gzip.BestSpeed)
		if err != nil {
			log.Printf("Failed to create gzip writer: %v", err)
			c.Next()
			return
		}
		defer func() {
			if closeErr := gz.Close(); closeErr != nil {
				log.Printf("failed to close gzip body: %s", closeErr.Error())
			}
		}()

		gzWriter := &GzipWriter{ResponseWriter: c.Writer, writer: gz}
		c.Writer = gzWriter

		c.Next()

		if err := gz.Flush(); err != nil {
			log.Printf("Failed to flush gzip writer: %v", err)
		}
	}
}
