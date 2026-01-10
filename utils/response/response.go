package response

import "github.com/gin-gonic/gin"

type Response struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Data    any       `json:"data"`
}

type ErrorCode int

const (
	ErrorCodeNone         ErrorCode = 0
	ErrorCodeGeneral      ErrorCode = 1
	ErrorCodeTokenExpired ErrorCode = 1000
	ErrorCodeTokenInvalid ErrorCode = 1001
)

func Error(c *gin.Context, message string, data any) {
	if len(message) == 0 {
		message = "fail"
	}
	c.JSON(200, Response{
		Code:    ErrorCodeGeneral,
		Message: message,
		Data:    data,
	})
}
func Success(c *gin.Context, data any) {
	c.JSON(200, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}
