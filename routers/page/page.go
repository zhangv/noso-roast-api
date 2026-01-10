package page

import (
	"hahacking/common/app"
	"hahacking/middleware"
	models "hahacking/models/page"
	"hahacking/utils"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

func RegisterRoutes(r *gin.RouterGroup) {
	APIs := r.Group("/page")
	{
		APIs.GET("meta/:ids", Meta)
		userAPI := APIs.Group("")
		userAPI.Use(middleware.JWTAuth())
		userAPI.Use(middleware.BaseAuth("user"))
		{
			userAPI.GET("", Search)
			userAPI.GET(":id", GetById)

		}
		adminAPI := APIs.Group("")
		adminAPI.Use(middleware.JWTAuth())
		adminAPI.Use(middleware.BaseAuth("admin"))
		{
			adminAPI.POST("", Update)
			adminAPI.DELETE(":id", Remove)
		}
	}
}

func Update(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var form models.UpdateForm
	if err := c.ShouldBindJSON(&form); err != nil {
		appG.Error(400, "绑定请求中的json到form失败，可能是格式问题", err)
		return
	}
	err := form.Validate()
	if len(err) > 0 {
		appG.Response(200, 1000, err)
		return
	}
	var o models.Page
	if err := mapstructure.Decode(form, &o); err != nil {
		appG.Error(400, "映射form内容到对象失败", err)
		return
	}
	o.ID = uint(form.ID)
	ro, err2 := models.Update(&o)

	if err2 != nil {
		appG.Error(400, "更新失败", err2)
		return
	}
	appG.Success("", ro)
}
func GetById(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.Atoi(c.Params.ByName("id"))
	o, err := models.GetById(uint(id))

	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	appG.Success("", o)
}

func Remove(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.Atoi(c.Params.ByName("id"))
	err = models.Delete(id)
	if err != nil {
		appG.Error(500, "删除失败", err)
		return
	}
	appG.Success("删除成功", "")
}
func Search(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var f models.SearchForm
	if c.ShouldBindQuery(&f) == nil {
		log.Println(f)
	}
	result, count := models.Search(f)
	appG.Success("", utils.Response{
		"objects": result,
		"total":   count,
	})
}

func Meta(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	name := c.Params.ByName("ids")
	result := models.Meta(name)
	appG.Success("", result)
}
