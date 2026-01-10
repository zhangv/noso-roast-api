package api

import (
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

type Form interface {
	Validate() map[string]string
}

func Bind(form Form, model any, c *gin.Context) map[string]string {
	var result = make(map[string]string)
	if err := c.ShouldBindJSON(form); err != nil {
		result["form"] = "字段绑定格式错误 - " + err.Error()
		return result
	}
	result = form.Validate()
	if len(result) > 0 {
		return result
	}

	if err := mapstructure.Decode(form, &model); err != nil {
		result["form"] = "映射form内容到对象失败"
		return result
	}

	return nil
}
