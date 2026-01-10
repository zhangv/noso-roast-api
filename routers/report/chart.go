package report

import (
	"hahacking/common/app"
	"hahacking/common/util"
	models "hahacking/models/report"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetChartData(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	chart, err := models.GetChart(c.Params.ByName("id"))
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	reqParams := util.GetQueryParams(c)
	data := chart.Plot(reqParams)
	appG.Success("", data)
}

func StatusChart(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	to := c.Params.ByName("to")
	err = models.StatusChart(id, to)
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	appG.Success("", nil)
}

func SearchChart(c *gin.Context) {
	var f models.SearchChartForm
	appG := app.Gin{Cxt: c}

	if c.ShouldBindQuery(&f) == nil {
		log.Println(f)
	}

	result, count := models.SearchChart(f)
	appG.Success("", map[string]any{
		"objects": result,
		"total":   count,
	})
}

func UpdateChart(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var f models.Chart
	c.ShouldBind(&f)
	errs := f.ValidateForm()
	if len(errs) > 0 {
		appG.Error(500, "表单报错", errs)
		return
	}
	log.Printf("chart %+v", f)
	err := models.UpdateChart(&f)
	if err != nil {
		appG.Error(500, "抱歉未找到相关信息", err.Error())
		return
	}
	appG.Success("", nil)
}
func GetChartById(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	chart, err := models.GetChart(c.Params.ByName("id"))
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	appG.Success("", chart)
}

func DeleteChart(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.Atoi(c.Params.ByName("id"))
	err = models.DeleteChart(id)

	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}

	appG.Success("", nil)
}
