package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shuklarituparn/go-metric-tracker/internal/utils"
)

func HMACMiddleware(key string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if key == "" {
			ctx.Next()
			return
		}
		if ctx.Request.Method == http.MethodPost {
			body, err := io.ReadAll(ctx.Request.Body)
			if err != nil {
				ctx.AbortWithStatus(http.StatusBadRequest)
				return
			}
			ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))
			recievedHash := ctx.GetHeader("HashSHA256")

			if len(body) > 0 && recievedHash != "" {
				calculatedHash := utils.CalculateHMAC(body, key)
				if calculatedHash != recievedHash {
					ctx.AbortWithStatus(http.StatusBadRequest)
					return
				}
			}
		}
		ctx.Next()

	}
}
