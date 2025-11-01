package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

type GzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *GzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

func (g *GzipWriter) WriteString(s string) (int, error) {
	return g.writer.Write([]byte(s))
}

func DecompressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Content-Encoding") == "gzip" {
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.AbortWithStatusJSON(400, gin.H{"error": "Failed to read request body"})
				return
			}
			if err := c.Request.Body.Close(); err != nil {
				log.Printf("failed to close request body: %s", err.Error())
			}

			reader, err := gzip.NewReader(bytes.NewReader(body))
			if err != nil {
				c.AbortWithStatusJSON(400, gin.H{"error": "Failed to decompress request body"})
				return
			}
			defer func() {
				if err := reader.Close(); err != nil {
					log.Printf("error closing reader: %s", err.Error())
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
			if err := gz.Close(); err != nil {
				log.Printf("failed to close gzip body: %s", err.Error())
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
