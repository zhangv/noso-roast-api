package roast

import (
	"fmt"
	"hahacking/common/app"
	models "hahacking/models/roast"
	"hahacking/utils"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 将烘焙参数保存为配方
func SaveRecipe(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	roastid, _ := strconv.Atoi(c.Param("id"))
	roast, err := models.GetById(uint(roastid))
	if err != nil {
		appG.Error(500, "获取烘焙信息失败,id="+strconv.Itoa(roastid), err)
		return
	}
	if roast.Status != models.RoastStatusComplete {
		appG.Error(500, fmt.Sprintf("烘焙状态[%s]未完成，不能保存配方", roast.Status), nil)
		return
	}
	recipe, err := models.SaveRecipeByRoastId(uint(roastid))
	if err != nil {
		appG.Error(500, "保存失败"+err.Error(), err)
		return
	}
	appG.Success("", recipe)
}
func SearchRecipe(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	o := map[string]any{}
	var f models.SearchRecipeForm

	if c.ShouldBindQuery(&f) == nil {
		log.Println(f)
	}

	if f.Name != "" {
		o["name"] = f.Name
	}
	if f.Status != "" {
		o["status"] = f.Status
	}
	o["pageNo"] = f.PageNo
	o["pageSize"] = f.PageSize

	result, count := models.SearchRecipe(f)
	appG.Success("", utils.Response{
		"objects": result,
		"total":   count,
	})
}
func UpdateRecipe(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var o models.RoastRecipe
	if err := c.ShouldBindJSON(&o); err != nil {
		appG.Error(400, "绑定请求中的json到form失败，可能是格式问题", err)
		return
	}
	err := models.UpdateRecipe(&o)
	if err != nil {
		appG.Error(500, "更新失败", err)
		return
	}
	appG.Success("", o)
}
func GetRecipeById(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, _ := strconv.Atoi(c.Param("id"))
	o, err := models.GetRecipeById(id)

	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	appG.Success("", o)
}
