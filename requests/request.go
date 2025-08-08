package requests

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/rolandhe/go-base/commons"
	"github.com/rolandhe/go-base/logger"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type BizFunc[T any, V any] func(ctx *commons.BaseContext, req *T) V

type RequestDesc[T, V any] struct {
	RelativePath  string
	AllowRoles    []string
	AllowProducts []int
	BizCoreFunc   BizFunc[T, V]
	LogLevel      LogLevel
	NotLogSQL     bool
}

func Get[T, V any](gg *gin.RouterGroup, rd *RequestDesc[T, V]) {
	gg.GET(rd.RelativePath, buildHandlersChain(rd)...)
}

func Post[T, V any](gg *gin.RouterGroup, rd *RequestDesc[T, V]) {
	gg.POST(rd.RelativePath, buildHandlersChain(rd)...)
}

func buildHandlersChain[T any, V any](rd *RequestDesc[T, V]) gin.HandlersChain {
	handlersChain := []gin.HandlerFunc{loginHandler(rd), doBizFunc(rd)}

	return handlersChain
}

func loginHandler[T any, V any](rd *RequestDesc[T, V]) gin.HandlerFunc {
	return func(gctx *gin.Context) {
		ctx := genBaseContext(gctx)
		url := gctx.Request.URL.Path

		if strings.Contains(url, "/public/") {
			token := commons.GetToken(ctx)
			err := PublicUserInfoCheckFunc(ctx, token, gctx.Request.URL.Path, ctx.QuickInfo())
			if err != nil {
				logger.WithBaseContextInfof(ctx)("get user info failed: %v", err)
				gctx.AbortWithStatusJSON(http.StatusOK, commons.QuickFromError(err))
				return
			}
			gctx.Next()
			return
		}

		if maybeShare(ctx) {
			gctx.Next()
			return
		}

		if strings.Contains(url, "/private/") {
			sUid := ctx.Get(commons.PrivateUid)
			if sUid != "" {
				uid, err := strconv.ParseInt(sUid, 10, 64)
				if err != nil {
					logger.WithBaseContextInfof(ctx)("parse private uid failed: %v", err)
					gctx.AbortWithStatusJSON(http.StatusOK, commons.QuickFromError(err))
					return
				}
				if err = PrivateUserInfoCheckFunc(ctx, uid, ctx.QuickInfo()); err != nil {
					logger.WithBaseContextInfof(ctx)("get private user info failed: %v", err)
					gctx.AbortWithStatusJSON(http.StatusOK, commons.QuickFromError(err))
					return
				}
				if ctx.QuickInfo().Uid == 0 {
					logger.WithBaseContextInfof(ctx)("not login in")
					gctx.AbortWithStatusJSON(http.StatusOK, NotLoginError)
					return
				}
			}
			gctx.Next()
			return
		}

		if strings.Contains(url, "/api/") {
			token := commons.GetToken(ctx)
			err := ApiUserInfoCheckFunc(ctx, token, gctx.Request.URL.Path, ctx.QuickInfo())
			if err != nil {
				logger.WithBaseContextInfof(ctx)("get user info failed: %v", err)
				gctx.AbortWithStatusJSON(http.StatusOK, commons.QuickFromError(err))
				return
			}
		}

		if ctx.QuickInfo().Uid == 0 {
			logger.WithBaseContextInfof(ctx)("not login in")
			gctx.AbortWithStatusJSON(http.StatusOK, NotLoginError)
			return
		}

		gctx.Next()
	}
}

func maybeShare(ctx *commons.BaseContext) bool {
	shareToken := ctx.Get(commons.ShareToken)
	return shareToken != ""
}

func doBizFunc[T any, V any](rd *RequestDesc[T, V]) gin.HandlerFunc {
	return func(gctx *gin.Context) {
		startUnixTs := time.Now().UnixMilli()

		ctx := genBaseContext(gctx)

		ctx.QuickInfo().NotLogSqlConf = rd.NotLogSQL

		var rt any
		reqObj := new(T)

		var bindFunc func(obj any) error

		// GET方法支持直接的query string，也支持form data,但x-www-form-urlencoded有问题
		if gctx.Request.Method == "GET" {
			bindFunc = gctx.ShouldBindQuery
		} else {
			bindFunc = func(obj any) error {
				return gctx.ShouldBindBodyWith(obj, binding.JSON)
			}
		}

		if err := bindFunc(reqObj); err != nil {
			beforeLog(gctx, ctx, rd.LogLevel)
			var errs validator.ValidationErrors
			if ok := errors.As(err, &errs); ok {
				logger.WithBaseContextInfof(ctx)("valid error")
				customErrMsgs := getCustomErrMsgs(reqObj)
				var errMsgs []string
				for _, e := range errs {
					ns := e.Namespace()
					customErrMsg, ok2 := customErrMsgs[ns]
					if ok2 {
						errMsgs = append(errMsgs, customErrMsg)
					} else {
						errMsgs = append(errMsgs, e.Error())
					}
				}

				msg := strings.Join(errMsgs, "\n")

				if msg != "" {
					rt = commons.QuickErrResult(msg)
				}
			} else {
				logger.WithBaseContextInfof(ctx)("bind request object error: %v", err)
			}
			if rt == nil {
				rt = commons.QuickErrResult("args invalid")
			}
		} else {
			beforeLog(gctx, ctx, rd.LogLevel)
			if maybeShare(ctx) {
				if err = ShareCheckFunc(ctx, reqObj, gctx.Request.URL.Path, ctx.QuickInfo()); err != nil {
					logger.WithBaseContextInfof(ctx)("check share token: %v", err)
					gctx.AbortWithStatusJSON(http.StatusOK, commons.QuickFromError(err))
					return
				}
			}
			rt = rd.BizCoreFunc(ctx, reqObj)
		}

		afterLog(ctx, rt, startUnixTs, rd.LogLevel)

		if !gctx.Writer.Written() {
			gctx.JSON(http.StatusOK, rt)
		}

		gctx.Next()
	}
}

func afterLog(baseCtx *commons.BaseContext, rt any, startUnixTs int64, ll LogLevel) {
	if ll == LOG_LEVEL_NONE {
		return
	}

	endUnixTs := time.Now().UnixMilli()
	latency := endUnixTs - baseCtx.GetCreateTime()
	bizCost := endUnixTs - startUnixTs

	uid := baseCtx.QuickInfo().Uid

	if ll&LOG_LEVEL_RETURN == LOG_LEVEL_RETURN {
		retJson, _ := json.Marshal(rt)
		logger.WithBaseContextInfof(baseCtx)("exit,uid=%d,ret is %s,cost=%d (%d) ms", uid, string(retJson), latency, bizCost)
		return
	}
	logger.WithBaseContextInfof(baseCtx)("exit,uid=%d,cost=%d (%d) ms", uid, latency, bizCost)
}

func beforeLog(gctx *gin.Context, baseCtx *commons.BaseContext, level LogLevel) {
	if level == LOG_LEVEL_NONE {
		return
	}
	uid := baseCtx.QuickInfo().Uid
	if level&LOG_LEVEL_PARAM == LOG_LEVEL_PARAM {
		bodyBytes, exists := gctx.Get(gin.BodyBytesKey)
		body := ""
		if exists {
			body = string(bodyBytes.([]byte))
		}
		keysContent := keysJson(gctx)
		logger.WithBaseContextInfof(baseCtx)("enter %s,uid=%d,keyHeader=%s,body is %s", gctx.Request.URL.String(), uid, keysContent, body)
		return
	}
	logger.WithBaseContextInfof(baseCtx)("enter %s,uid=%d", gctx.Request.URL.String(), uid)
}

func getCustomErrMsgs(req any) map[string]string {
	reqType := reflect.TypeOf(req)
	if reqType.Kind() != reflect.Ptr || reqType.Elem().Kind() != reflect.Struct {
		return map[string]string{}
	}

	errMsgs := map[string]string{}
	findCustomErrMsgs(reqType.Elem(), reqType.Elem().Name(), "", errMsgs)
	return errMsgs
}

func findCustomErrMsgs(tType reflect.Type, tName string, tPath string, errMsgs map[string]string) {
	var sType reflect.Type
	sTypeKind := tType.Kind()
	if sTypeKind == reflect.Ptr {
		sType = tType.Elem()
	} else {
		sType = tType
	}

	tPath = tPath + tName + "."

	for i := 0; i < sType.NumField(); i++ {
		f := sType.Field(i)
		fType := f.Type
		fName := f.Name

		errMsg := f.Tag.Get("errMsg")
		if errMsg != "" {
			ns := tPath + fName
			errMsgs[ns] = errMsg
		}

		if fType.Kind() == reflect.Ptr && fType.Elem().Kind() == reflect.Struct {
			findCustomErrMsgs(fType, fName, tPath, errMsgs)
		}
	}
}

var keyHeaders = []string{
	"device-id",
	"hardware",
	"os",
	"os-version",
	"app-version",
}

func keysJson(gctx *gin.Context) string {
	km := map[string]string{}
	for _, k := range keyHeaders {
		v := getHeader(gctx, k)
		km[k] = v
	}
	j, _ := json.Marshal(km)
	return string(j)
}
