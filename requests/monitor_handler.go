package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/rolandhe/go-base/monitor"
	"time"
)

func monitorHandler() gin.HandlerFunc {
	return func(gctx *gin.Context) {
		path := gctx.Request.URL.Path
		monitor.DoServerCounter(path)
		start := time.Now().UnixMilli()
		gctx.Next()

		cost := time.Now().UnixMilli() - start
		e := "false"
		if len(gctx.Errors) > 0 {
			e = "true"
		}
		monitor.DoServerDuration(path, e, cost)
	}
}
