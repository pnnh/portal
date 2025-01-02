package helpers

import "github.com/gin-gonic/gin"

func GetIpAddress(gctx *gin.Context) string {
	ip := gctx.GetHeader("X-Real-IP")
	if ip == "" {
		ip = gctx.GetHeader("X-Forwarded-For")
	}
	if ip == "" {
		ip = gctx.GetHeader("cf-connecting-ip")
	}
	if ip == "" {
		ip = gctx.ClientIP()
	}
	return ip
}
