package middleware

import (
	"bytes"
	"crypto/rsa"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shuklarituparn/go-metric-tracker/internal/utils"
)

// DecryptionMiddleware создает middleware для расшифровки данных.
// Расшифровывает тело запроса если заголовок Content-Encrypted равен "rsa".
// Параметры:
//   - privateKey: приватный RSA ключ для расшифровки (если nil, middleware пропускает запрос)
func DecryptionMiddleware(privateKey *rsa.PrivateKey) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if privateKey == nil {
			ctx.Next()
			return
		}

		if ctx.Request.Method == http.MethodPost {
			contentEncrypted := ctx.GetHeader("Content-Encrypted")
			if contentEncrypted == "rsa" {
				body, err := io.ReadAll(ctx.Request.Body)
				if err != nil {
					log.Printf("Failed to read encrypted body: %v", err)
					ctx.AbortWithStatus(http.StatusBadRequest)
					return
				}

				decryptedData, err := utils.DecryptData(body, privateKey)
				if err != nil {
					log.Printf("Failed to decrypt data: %v", err)
					ctx.AbortWithStatus(http.StatusBadRequest)
					return
				}

				ctx.Request.Body = io.NopCloser(bytes.NewBuffer(decryptedData))
				ctx.Request.ContentLength = int64(len(decryptedData))

				ctx.Request.Header.Del("Content-Encrypted")
			}
		}
		ctx.Next()
	}
}