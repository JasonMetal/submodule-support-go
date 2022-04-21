package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// ------------请求信息基础支撑----------

type Request struct {
	GCtx *gin.Context
	Val  string
}

type Number interface {
	~int
}

func NewRequest(ctx *gin.Context) Request {
	return Request{GCtx: ctx}
}

// GetHeader 获取响应头
func (req Request) GetHeader(key string) Request {
	req.Val = req.GCtx.GetHeader(key)
	return req
}

func (req Request) GetQuery(key string) Request {
	req.Val = req.GCtx.Query(key)
	return req
}

func (req Request) GetQueryDefault(key string, defaultValue string) Request {
	req.Val = req.GCtx.DefaultQuery(key, defaultValue)
	return req
}

func (req Request) PostForm(key string) Request {
	req.Val = req.GCtx.PostForm(key)
	return req
}

func (req Request) Value() string {
	return req.Val
}

func (req Request) Bool() bool {
	b, _ := strconv.ParseBool(req.Val)
	return b
}

func (req Request) ShouldBindQuery(obj interface{}) {
	_ = req.GCtx.ShouldBindQuery(obj)
}

// ShouldBindJSON 使用注意，只能获取一次，无法重复获取
// ShouldBindWith for better performance if you need to call only once.
func (req Request) ShouldBindJSON(obj interface{}) {
	_ = req.GCtx.ShouldBindJSON(obj)
}

func (req Request) IsMimeJson() bool {
	return req.GCtx.ContentType() == binding.MIMEJSON
}

func (req Request) PostToModel(obj interface{}) Request {
	if req.IsMimeJson() {
		_ = req.GCtx.ShouldBindBodyWith(obj, binding.JSON)
	} else {
		_ = req.GCtx.ShouldBind(obj)
	}
	return req
}

// GetAllParamsFromUrl 获取url中所有参数, 不支持数组
func (req Request) GetAllParamsFromUrl() map[string]string {
	reqParams := make(map[string]string)
	params := req.GCtx.Request.URL.Query()
	for key, value := range params {
		if len(value) == 1 {
			reqParams[key] = value[0]
		}
	}

	return reqParams
}
