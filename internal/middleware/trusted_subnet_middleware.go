package middleware

import (
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)


func TrustedSubnetMiddleware(trustedSubnet string) gin.HandlerFunc {
	if trustedSubnet == "" {
		return func(ctx *gin.Context) {
			ctx.Next()
		}
	}

	_, subnet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		log.Printf("Warning: invalid trusted subnet CIDR %q: %v. All requests will be forbidden.", trustedSubnet, err)
		return func(ctx *gin.Context) {
			ctx.AbortWithStatus(http.StatusForbidden)
		}
	}

	return func(ctx *gin.Context) {
		realIP := ctx.GetHeader("X-Real-IP")
		if realIP == "" {
			log.Printf("Request rejected: missing X-Real-IP header")
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		ip := net.ParseIP(realIP)
		if ip == nil {
			log.Printf("Request rejected: invalid IP address in X-Real-IP header: %s", realIP)
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		if !subnet.Contains(ip) {
			log.Printf("Request rejected: IP %s is not in trusted subnet %s", realIP, trustedSubnet)
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		ctx.Next()
	}
}