package shop

import (
	"hahacking/common/app"
	"hahacking/common/logger"
	models "hahacking/models/shop"
	"hahacking/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetAccountLogById(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, _ := strconv.Atoi(c.Params.ByName("id"))
	o, err := models.GetAccountLogById(id)

	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	appG.Success("", o)
}

func SearchAccountLog(c *gin.Context) {
	var f models.SearchAccountLogForm
	appG := app.Gin{Cxt: c}

	if c.ShouldBindQuery(&f) == nil {
		logger.Infof("SearchAccountLog: %+v", f)
	}

	result, count := models.SearchAccountLog(f)
	appG.Success("", utils.Response{
		"objects": result,
		"total":   count,
	})
}
