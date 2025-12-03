package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// DefaultHandle обрабатывает запросы к корневому пути приложения.
// Маршрут: GET /
// Возвращает HTML страницу с приветственным сообщением.
func DefaultHandle(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/html")
	ctx.String(http.StatusOK,
		"<html><body>"+strings.Repeat("Hello, world<br>", 20)+"</body></html>")
}
