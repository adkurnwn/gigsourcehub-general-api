package middleware

import (
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/gin-gonic/gin"
)

func (c *appMiddleware) ContextEnricher() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 1. Robust IP Extraction (Cloudflare Aware)
		ipAddress := ctx.ClientIP()
		if cfIP := ctx.GetHeader("CF-Connecting-IP"); cfIP != "" {
			ipAddress = cfIP
		}

		// 2. Extra Metadata
		userAgent := ctx.Request.UserAgent()
		endpoint := ctx.FullPath()

		// 3. Enrich Request Context
		rCtx := ctx.Request.Context()
		rCtx = helpers.SetIPAddress(rCtx, ipAddress)
		rCtx = helpers.SetUserAgent(rCtx, userAgent)
		rCtx = helpers.SetEndpoint(rCtx, endpoint)

		// 4. Update Request with Enriched Context
		ctx.Request = ctx.Request.WithContext(rCtx)

		ctx.Next()
	}
}
