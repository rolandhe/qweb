package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/rolandhe/qweb/profile"
)

const MaxMultipartMemory = 2 << 20

func NewEngine() *gin.Engine {
	if profile.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}
	e := gin.New()
	e.UseH2C = true
	e.MaxMultipartMemory = MaxMultipartMemory
	e.Use(recoverHandler(), corsHandler(), monitorHandler(), healthHandler())
	_ = e.SetTrustedProxies(nil)
	e.HandleMethodNotAllowed = true

	return e
}
