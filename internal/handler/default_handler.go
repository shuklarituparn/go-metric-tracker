package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func DefaultHandle(ctx *gin.Context) {
    ctx.Header("Content-Type", "text/html")
    ctx.String(http.StatusOK,
        "<html><body>"+strings.Repeat("Hello, world<br>", 20)+"</body></html>")
}
