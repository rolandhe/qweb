package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/rolandhe/go-base/logger"
	"net/http"
)

func healthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.RequestURI == "/health" {
			baseContext := genBaseContext(c)
			logger.WithBaseContextInfof(baseContext)("health check")
			c.AbortWithStatusJSON(http.StatusOK, "ok")
		}
	}
}
