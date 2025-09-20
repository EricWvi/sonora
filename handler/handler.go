package handler

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/EricWvi/sonora/log"
	"github.com/gin-gonic/gin"
)

type Response struct {
	RequestId string `json:"requestId"`
	Code      int    `json:"code"`
	Message   any    `json:"message"`
}

func Dispatch(c *gin.Context, base any) {
	method := reflect.ValueOf(base).MethodByName(c.GetString("Action"))
	if !method.IsValid() {
		ReplyError(c, http.StatusNotFound, "request action does not exist")
		return
	}
	ctx := reflect.ValueOf(c)
	Type := method.Type()
	param := Type.In(1).Elem()
	ptr := reflect.New(param).Interface()

	// Let Gin handle binding (body, query)
	if err := c.ShouldBind(ptr); err != nil {
		ReplyError(c, http.StatusBadRequest, "failed to bind request: "+err.Error())
		return
	}

	rst := method.Call([]reflect.Value{ctx, reflect.ValueOf(ptr)})[0]
	if !c.IsAborted() {
		c.JSON(http.StatusOK, Response{
			RequestId: c.GetString(log.RequestIDCtxKey),
			Code:      http.StatusOK,
			Message:   rst.Interface(),
		})
	}
}

func ReplyError(c *gin.Context, code int, msg string) {
	c.JSON(code, Response{
		RequestId: c.GetString(log.RequestIDCtxKey),
		Code:      code,
		Message:   msg,
	})
}

func ReplyString(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		RequestId: c.GetString(log.RequestIDCtxKey),
		Code:      code,
		Message:   msg,
	})
}

func Errorf(c *gin.Context, format string, a ...any) {
	ReplyString(c, http.StatusBadRequest, fmt.Sprintf(format, a...))
	c.Abort()
}
