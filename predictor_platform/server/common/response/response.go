package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

const (
	SUCCESS = 0
)

func Result(code int, data interface{}, msg string, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		code,
		data,
		msg,
	})
}

func Done(c *gin.Context) {
	Result(SUCCESS, map[string]interface{}{}, "Done", c)
}

func DoneWithMessage(message string, c *gin.Context) {
	Result(SUCCESS, map[string]interface{}{}, message, c)
}

func DoneWithData(data interface{}, c *gin.Context) {
	Result(SUCCESS, data, "Done", c)
}

func DoneWithDetail(data interface{}, message string, c *gin.Context) {
	Result(SUCCESS, data, message, c)
}

func ResultWithoutData(code int, message string, c *gin.Context) {
	Result(code, map[string]interface{}{}, message, c)
}
