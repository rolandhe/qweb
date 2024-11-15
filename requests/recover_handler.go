package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/rolandhe/go-base/commons"
	"github.com/rolandhe/go-base/logger"
	"github.com/rolandhe/qweb/profile"
	"go/types"
	"net/http"
)

func recoverHandler() gin.HandlerFunc {
	return gin.CustomRecovery(myRecover)
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
		if profile.IsProd() {
			return commons.QuickErrResult("internal server error")
		} else {
			return commons.QuickErrResult(v.Error())
		}
	default:
		return commons.QuickErrResult("unknown error")
	}
}
