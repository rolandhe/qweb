package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/rolandhe/go-base/commons"
	"github.com/rolandhe/qweb/cors"
)

var AllowOrigins = []string{
	"*",
}

var AllowHeaders = []string{
	commons.TraceId,
	commons.Platform,
	commons.Token,
	commons.ShareToken,

	"device-id",
	"hardware",
	"os",
	"os-version",
	"resolution",
	"app-key",
	"app-version",
	"app_vsn",
	"X-Forwarded-For",
	"X-Forwarded-Proto",
	"Authorization",
}

func corsHandler() gin.HandlerFunc {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = AllowOrigins
	corsConfig.AllowHeaders = AllowHeaders

	return cors.New(corsConfig)
}
