package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/rolandhe/go-base/commons"
	"github.com/rolandhe/go-base/envsupport"
	"github.com/rolandhe/go-base/logger"
	"go/types"
	"net/http"
	"strings"
)

func recoverHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		out := &strings.Builder{}
		f := gin.RecoveryWithWriter(out, myRecover)
		f(c)
		baseCtx := genBaseContext(c)
		logger.WithBaseContextErrorf(baseCtx)("panic stack: %s", out.String())
	}
}

func myRecover(gctx *gin.Context, err any) {
	baseCtx := genBaseContext(gctx)
	logger.WithBaseContextErrorf(baseCtx)("panic error: %v", err)

	gctx.JSON(http.StatusOK, errorToResult(err))
	gctx.Abort()
}

func errorToResult(r any) any {
	switch v := r.(type) {
	case string:
		return commons.QuickErrResult(v)
	case *commons.StdError:
		return commons.NewResult[*types.Nil](v.Code, v.Message, nil)
	case error:
		if envsupport.Profile() == "prod" {
			return commons.QuickErrResult("internal server error")
		} else {
			return commons.QuickErrResult(v.Error())
		}
	default:
		return commons.QuickErrResult("unknown error")
	}
}
