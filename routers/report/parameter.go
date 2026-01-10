package report

import (
	"hahacking/common/app"
	models "hahacking/models/report"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

func StatusParameter(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	to := c.Params.ByName("to")
	err = models.StatusParameter(id, to)
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	appG.Success("", nil)
}

func SearchParameter(c *gin.Context) {
	var f models.SearchParameterForm
	appG := app.Gin{Cxt: c}

	if c.ShouldBindQuery(&f) == nil {
		log.Println(f)
	}

	result, count := models.SearchParameter(f)
	appG.Success("", map[string]any{
		"objects": result,
		"total":   count,
	})
}

func UpdateParameter(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var f models.Parameter
	c.ShouldBindJSON(&f)
	errs := f.ValidateForm()
	if len(errs) > 0 {
		appG.Error(500, "表单报错", errs)
		return
	}
	log.Printf("parameter %+v", f)
	err := models.UpdateParameter(&f)
	if err != nil {
		appG.Error(500, "抱歉未找到相关信息", err.Error())
		return
	}
	appG.Success("", nil)
}
func GetParameterById(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.Atoi(c.Params.ByName("id"))
	parameter, err := models.GetParameterById(uint(id))
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	appG.Success("", parameter)
}

func DeleteParameter(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.Atoi(c.Params.ByName("id"))
	err = models.DeleteChart(id)

	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}

	appG.Success("", nil)
}
