package http

import (
	"bytes"
	"context"
	"errors"
	broccolictx "github.com/elvisNg/broccoli/context"
	"github.com/elvisNg/broccoli/engine"
	broccolierrors "github.com/elvisNg/broccoli/errors"
	"github.com/elvisNg/broccoli/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/jsonpb"
	proto "github.com/golang/protobuf/proto"
	"github.com/opentracing/opentracing-go"
	zipkintracer "github.com/openzipkin/zipkin-go-opentracing"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"reflect"
)

const BROCCOLI_CTX = "broccolictx"
const BROCCOLI_HTTP_ERR = "broccoli_http_err"
const BROCCOLI_HTTP_REWRITE_ERR = "broccoli_http_rewrite_err"
const BROCCOLI_HTTP_REWRITE_RESPONSE = "broccoli_http_rewrite_response"
const BROCCOLI_HTTP_DISABLE_PB_VALIDATE = "broccoli_http_disable_pb_validate"
const BROCCOLI_HTTP_USE_GINBIND_VALIDATE_FOR_PB = "broccoli_http_use_ginbind_validate_for_pb"

var SuccessResponse SuccessResponseHandler = defaultSuccessResponse
var ErrorResponse ErrorResponseHandler = defaultErrorResponse

type SuccessResponseHandler func(c *gin.Context, rsp interface{})
type ErrorResponseHandler func(c *gin.Context, err error)
type reWriteErrFn func(c *gin.Context, err error)

type validator interface {
	Validate() error
}

// SetReWriteErrFn 自定义错误处理
func SetReWriteErrFn(f reWriteErrFn) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Set(BROCCOLI_HTTP_REWRITE_ERR, f)
		c.Next()
	}
}

type reWriteResponseFn func(c *gin.Context, rsp interface{})

// SetReWriteResponseFn 自定义返回处理
func SetReWriteResponseFn(f reWriteResponseFn) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Set(BROCCOLI_HTTP_REWRITE_RESPONSE, f)
		c.Next()
	}
}

func NotFound(ng engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		ExtractLogger(c).Debugf("url not found url: %s\n", c.Request.URL)
		// c.JSON(http.StatusNotFound, "not found")
		c.String(http.StatusNotFound, "not found")
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func Access(ng engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := ng.GetContainer().GetLogger()
		ctx := c.Request.Context()
		l := logger.WithFields(logrus.Fields{"tag": "gin"})
		////// zipkin begin
		cfg, err := ng.GetConfiger()
		if err != nil {
			l.Error(err)
			return
		}
		name := c.Request.URL.Path
		tracer := ng.GetContainer().GetTracer()
		if tracer == nil {
			l.Error("tracer is nil")
			return
		}
		spnctx, span, err := tracer.StartSpanFromContext(ctx, name)
		if err != nil {
			l.Error(err)
			return
		}

		header, _ := utils.Marshal(c.Request.Header)
		span.SetTag("http request.header", string(header))
		span.SetTag("http request.method", c.Request.Method)
		span.SetTag("http request.url", c.Request.URL.String())

		if c.Request.Body != nil {
			bodyBytes, err := ioutil.ReadAll(c.Request.Body)
			if err == nil {
				span.SetTag("http request.body", string(bodyBytes))
				// Restore the io.ReadCloser to its original state
				c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		if cfg.Get().Trace.OnlyLogErr {
			c.Writer = blw
		}

		// before request
		defer func() {
			if blw.body.Len() > 0 && blw.body.Bytes()[0] == '{' {
				baseRsp := struct {
					Errcode int32 `json:"errcode"`
				}{}
				if err1 := utils.Unmarshal(blw.body.Bytes(), &baseRsp); err1 == nil && baseRsp.Errcode == 0 {
					return
				}
			}
			span.Finish()
		}()
		////// zipkin finish
		l = l.WithFields(logrus.Fields{"tracerid": span.Context().(zipkintracer.SpanContext).TraceID.ToHex()})
		ctx = broccolictx.LoggerToContext(spnctx, l)
		ctx = broccolictx.EngineToContext(ctx, ng)
		ctx = broccolictx.GMClientToContext(ctx, ng.GetContainer().GetGoMicroClient())
		if ng.GetContainer().GetRedisCli() != nil {
			ctx = broccolictx.RedisToContext(ctx, ng.GetContainer().GetRedisCli().GetCli())
		}
		if ng.GetContainer().GetMysql() != nil {
			ctx = broccolictx.MysqlToContext(ctx, ng.GetContainer().GetMysql())
		}
		c.Set(BROCCOLI_CTX, ctx)
		l.Debugln("access start", c.Request.URL.Path)
		c.Next()
		l.Debugln("access end", c.Request.URL.Path)
	}
}

func ExtractLogger(c *gin.Context) *logrus.Entry {
	ctx := c.Request.Context()
	if cc, ok := c.Value(BROCCOLI_CTX).(context.Context); ok && cc != nil {
		ctx = cc
	}
	return broccolictx.ExtractLogger(ctx)
}

func ExtractTracerID(c *gin.Context) string {
	ctx := c.Request.Context()
	if cc, ok := c.Value(BROCCOLI_CTX).(context.Context); ok && cc != nil {
		ctx = cc
	}
	span := opentracing.SpanFromContext(ctx)
	return span.Context().(zipkintracer.SpanContext).TraceID.ToHex()
}

func ExtractEngine(c *gin.Context) (engine.Engine, error) {
	ctx := c.Request.Context()
	if cc, ok := c.Value(BROCCOLI_CTX).(context.Context); ok && cc != nil {
		ctx = cc
	}
	return broccolictx.ExtractEngine(ctx)
}

func defaultSuccessResponse(c *gin.Context, rsp interface{}) {
	logger := ExtractLogger(c)
	logger.Debug("defaultSuccessResponse")
	res := broccolierrors.New(broccolierrors.ECodeSuccessed, "", "")
	res.TracerID = ExtractTracerID(c)
	if ng, _ := ExtractEngine(c); ng != nil {
		res.ServiceID = ng.GetContainer().GetServiceID()
	}
	res.Data = rsp
	f, exists := c.Get(BROCCOLI_HTTP_REWRITE_RESPONSE)
	if exists && f != nil {
		c.Set(BROCCOLI_HTTP_REWRITE_RESPONSE, nil)
		ff, ok := f.(reWriteResponseFn)
		if ok {
			ff(c, rsp)
			return
		}
	}
	res.Write(c.Writer)
}

func defaultErrorResponse(c *gin.Context, err error) {
	logger := ExtractLogger(c)
	logger.Debug("defaultErrorResponse")
	broccoliErr := assertError(err)
	if broccoliErr == nil {
		broccoliErr = broccolierrors.New(broccolierrors.ECodeSystem, "err was a nil error or was a nil *broccolierrors.Error", "assertError")
	}
	if utils.IsEmptyString(broccoliErr.TracerID) {
		broccoliErr.TracerID = ExtractTracerID(c)
	}
	if utils.IsEmptyString(broccoliErr.ServiceID) {
		if ng, _ := ExtractEngine(c); ng != nil {
			broccoliErr.ServiceID = ng.GetContainer().GetServiceID()
		}
	}
	c.Set(BROCCOLI_HTTP_ERR, err)
	f, exists := c.Get(BROCCOLI_HTTP_REWRITE_ERR)
	if exists && f != nil {
		c.Set(BROCCOLI_HTTP_REWRITE_ERR, nil)
		ff, ok := f.(reWriteErrFn)
		if ok {
			ff(c, err)
			return
		}
	}
	broccoliErr.Write(c.Writer)
}

func assertError(e error) (err *broccolierrors.Error) {
	if e == nil {
		return
	}
	if utils.IsBlank(reflect.ValueOf(e)) {
		return
	}
	var broccoliErr *broccolierrors.Error
	if errors.As(e, &broccoliErr) {
		err = broccoliErr
		return
	}
	err = broccolierrors.New(broccolierrors.ECodeSystemAPI, e.Error(), "assertError")
	return
}

func GenerateGinHandle(handleFunc interface{}) func(c *gin.Context) {
	return func(c *gin.Context) {
		h := reflect.ValueOf(handleFunc)
		reqT := h.Type().In(1).Elem()
		rspT := h.Type().In(2).Elem()

		reqV := reflect.New(reqT)
		rspV := reflect.New(rspT)

		req := reqV.Interface()
		// 针对proto.Message进行反序列化和校验
		if pb, ok := req.(proto.Message); ok {
			if c.GetBool(BROCCOLI_HTTP_USE_GINBIND_VALIDATE_FOR_PB) {
				if err := c.ShouldBind(req); err != nil {
					ExtractLogger(c).Error(err)
					ErrorResponse(c, broccolierrors.ECodeInvalidParams.ParseErr(err.Error()))
					return
				}
			} else {
				val := make(map[string]interface{})
				if err := c.ShouldBind(&val); err != nil {
					ExtractLogger(c).Debug(err)
					ErrorResponse(c, broccolierrors.ECodeJsonUnmarshal.ParseErr(err.Error()))
					return
				}
				b, err := utils.Marshal(val)
				if err != nil {
					ExtractLogger(c).Debug(err)
					ErrorResponse(c, broccolierrors.ECodeJsonMarshal.ParseErr(err.Error()))
					return
				}
				// 允许未定义pb
				m := jsonpb.Unmarshaler{
					AllowUnknownFields: true,
				}
				err = m.Unmarshal(bytes.NewBuffer(b), pb)
				if err != nil {
					ExtractLogger(c).Debug(err)
					ErrorResponse(c, broccolierrors.ECodeJSONPBUnmarshal.ParseErr(err.Error()))
					return
				}
				if !c.GetBool(BROCCOLI_HTTP_DISABLE_PB_VALIDATE) {
					if v, ok := req.(validator); v != nil && ok {
						if err := v.Validate(); err != nil {
							ExtractLogger(c).Debug(err)
							ErrorResponse(c, broccolierrors.ECodeInvalidParams.ParseErr(err.Error()))
							return
						}
					}
				}
			}
		} else {
			// 非proto.Message
			if err := c.ShouldBind(req); err != nil {
				ExtractLogger(c).Debug(err)
				ErrorResponse(c, broccolierrors.ECodeInvalidParams.ParseErr(err.Error()))
				return
			}

		}

		ctx := c.Request.Context()
		if cc, ok := c.Value(BROCCOLI_CTX).(context.Context); ok && cc != nil {
			ctx = cc
		}
		ctxV := reflect.ValueOf(ctx)
		ret := h.Call([]reflect.Value{ctxV, reqV, rspV})
		if !ret[0].IsNil() {
			err, ok := ret[0].Interface().(error)
			if ok {
				ErrorResponse(c, err)
				return
			}
			ErrorResponse(c, broccolierrors.ECodeInternal.ParseErr("UNKNOW ERROR"))
			return
		}
		SuccessResponse(c, rspV.Interface())
	}
}
