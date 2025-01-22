package requests

import (
	"github.com/gin-gonic/gin"
)

const MaxMultipartMemory = 2 << 20

func NewEngine(ginMode string) *gin.Engine {
	gin.SetMode(ginMode)
	e := gin.New()
	e.UseH2C = true
	e.MaxMultipartMemory = MaxMultipartMemory
	e.Use(recoverHandler(), corsHandler(), monitorHandler(), healthHandler())
	_ = e.SetTrustedProxies(nil)
	e.HandleMethodNotAllowed = true

	return e
}
