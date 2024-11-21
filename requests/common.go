package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rolandhe/go-base/commons"
)

const (
	baseContextName = "base_context_qweb"
)

var ApiUserInfoCheckFunc = func(ctx *commons.BaseContext, token string, platform string, urlPath string, info *commons.QuickInfo) error {
	panic("请设置ApiUserInfoCheckFunc")
}
var ShareCheckFunc = func(ctx *commons.BaseContext, queryParam map[string]string, info *commons.QuickInfo) error {
	panic("请设置ShareCheckFunc")
}
var PrivateUserInfoCheckFunc = func(ctx *commons.BaseContext, uid int64, info *commons.QuickInfo) error {
	return nil
}

func genBaseContext(gctx *gin.Context) *commons.BaseContext {
	v, exists := gctx.Get(baseContextName)
	if exists {
		return v.(*commons.BaseContext)
	}

	baseContext := commons.NewBaseContext()
	tid := getHeader(gctx, commons.TraceId)
	if tid == "" {
		tid = uuid.NewString() + "-cr"
	}
	baseContext.Put(commons.TraceId, tid)
	baseContext.Put(commons.Profile, getHeader(gctx, commons.Profile))
	baseContext.Put(commons.Token, getToken(gctx))
	baseContext.Put(commons.Platform, getHeader(gctx, commons.Platform))
	baseContext.Put(commons.ShareToken, getShareToken(gctx))
	privateUid := getHeader(gctx, commons.PrivateUid)
	if privateUid != "" {
		baseContext.Put(commons.PrivateUid, privateUid)
	}

	gctx.Set(baseContextName, baseContext)
	return baseContext
}

func getToken(gctx *gin.Context) string {
	token := getHeader(gctx, commons.Token)
	if len(token) == 0 {
		token = gctx.Query(commons.Token)
	}
	return token
}

func getShareToken(gctx *gin.Context) string {
	token := getHeader(gctx, commons.ShareToken)
	if len(token) == 0 {
		token = gctx.Query(commons.ShareToken)
	}
	return token
}

func getHeader(gctx *gin.Context, key string) string {
	header := gctx.GetHeader(key)
	if header != "" {
		return header
	}

	headers := gctx.Request.Header[key]
	if len(headers) > 0 {
		return headers[0]
	}

	return ""
}
