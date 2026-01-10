package app

import (
	"hahacking/common/ret"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Gin struct {
	Cxt *gin.Context
}

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"message"`
	Data any    `json:"data"`
}

// Response setting gin.JSON
func (g *Gin) Response(httpCode, errCode int, data any) {
	g.Cxt.JSON(httpCode, Response{
		Code: errCode,
		Msg:  ret.GetMsg(errCode),
		Data: data,
	})
}

func (g *Gin) Success(msg string, data any) {
	g.Cxt.JSON(http.StatusOK, gin.H{
		"code":    ret.SUCCESS,
		"message": msg,
		"data":    data,
	})
}

func (g *Gin) SuccessWithNoData(msg string) {
	g.Cxt.JSON(http.StatusOK, gin.H{
		"code":    ret.SUCCESS,
		"message": msg,
	})
}

func (g *Gin) Failed(code int, msg string) {
	g.Cxt.AbortWithStatusJSON(http.StatusOK, gin.H{
		"code":    code,
		"message": msg,
	})
}

func (g *Gin) Error(code int, msg string, data interface{}) {
	g.Cxt.AbortWithStatusJSON(http.StatusOK, gin.H{
		"code":    code,
		"message": msg,
		"data":    data,
	})
}

func (g *Gin) UnAuthorized() {
	g.Cxt.JSON(http.StatusOK, gin.H{
		"code":    ret.TOKEN_INVALID,
		"message": ret.GetMsg(ret.TOKEN_INVALID),
	})
	g.Cxt.Abort()
}
func (g *Gin) NotFound(message string) {
	g.Cxt.JSON(http.StatusNotFound, gin.H{
		// "code":    ret.TOKEN_INVALID,
		"message": message,
	})
	g.Cxt.Abort()
}
