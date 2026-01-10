package report

import (
	"hahacking/common/app"
	models "hahacking/models/report"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

func StatusColumn(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	to := c.Params.ByName("to")
	err = models.StatusColumn(id, to)
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	appG.Success("", nil)
}

func SearchColumn(c *gin.Context) {
	var f models.SearchColumnForm
	appG := app.Gin{Cxt: c}

	if c.ShouldBindQuery(&f) == nil {
		log.Println(f)
	}

	result, count := models.SearchColumn(f)
	appG.Success("", map[string]any{
		"objects": result,
		"total":   count,
	})
}

func UpdateColumn(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var f models.Column
	c.ShouldBind(&f)
	errs := f.ValidateForm()
	if len(errs) > 0 {
		appG.Error(500, "表单报错", errs)
		return
	}
	log.Printf("column %+v", f)
	err := models.UpdateColumn(&f)
	if err != nil {
		appG.Error(500, "抱歉未找到相关信息", err.Error())
		return
	}
	appG.Success("", nil)
}
func GetColumnById(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.Atoi(c.Params.ByName("id"))
	column, err := models.GetColumnById(uint(id))
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	appG.Success("", column)
}

func DeleteColumn(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.Atoi(c.Params.ByName("id"))
	err = models.DeleteColumn(id)

	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}

	appG.Success("", nil)
}
