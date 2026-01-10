package roast

import (
	"hahacking/common/app"
	"hahacking/middleware"
	models "hahacking/models/roast"
	"hahacking/utils"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	APIs := r.Group("/roast")
	{
		APIs.GET("meta/:ids", Meta)
	}
	userAPI := APIs.Group("")
	userAPI.Use(middleware.JWTAuth())
	userAPI.Use(middleware.BaseAuth("user"))
	{
		userAPI.GET("", Search)
		userAPI.GET(":id", GetById)
		userAPI.GET(":id/profile", GetProfileByRoastId)
		userAPI.POST(":id/saverecipe", SaveRecipe)
		userAPI.POST("start", Start)
		userAPI.POST("complete", Complete)
		userAPI.GET(":id/cancel", Cancel)
		userAPI.GET(":id/delete", Delete)

		userAPI.GET("profile/:id", GetProfileById)
		userAPI.POST("profile", UpdateProfile)

		userAPI.GET("recipe", SearchRecipe)
		userAPI.GET("recipe/:id", GetRecipeById)
		userAPI.POST("recipe", UpdateRecipe)
	}

	adminAPI := APIs.Group("")
	adminAPI.Use(middleware.JWTAuth())
	adminAPI.Use(middleware.BaseAuth("admin"))
	{

	}
}
func Meta(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	name := c.Params.ByName("ids")
	result := models.Meta(name)
	appG.Success("", result)
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

func Search(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	o := map[string]any{}
	var f models.SearchForm

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

	result, count := models.Search(f)
	appG.Success("", utils.Response{
		"objects": result,
		"total":   count,
	})
}

func Start(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var form models.StartForm
	if err := c.ShouldBindJSON(&form); err != nil {
		appG.Error(400, "绑定请求中的json到form失败，可能是格式问题", err)
		return
	}
	formErrs := form.Validate()
	if len(formErrs) > 0 {
		appG.Response(200, 1000, formErrs)
		return
	}
	roast, err := models.Start(form)
	if err != nil {
		appG.Error(500, "更新失败", err)
		return
	}
	appG.Success("", roast)
}

func Complete(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var form models.CompleteForm
	if err := c.ShouldBindJSON(&form); err != nil {
		appG.Error(400, "绑定请求中的json到form失败，可能是格式问题", err)
		return
	}
	formErrs := form.Validate()
	if len(formErrs) > 0 {
		appG.Response(200, 1000, formErrs)
		return
	}
	roast, _ := models.GetById(form.ID)
	err := roast.Complete(form)
	if err != nil {
		log.Println("complete error", err)
		appG.Error(500, "完成失败", err)
		return
	}
	roast, _ = models.GetById(form.ID)
	appG.Success("", roast)
}

func Cancel(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, _ := strconv.Atoi(c.Params.ByName("id"))
	err := models.Cancel(id)

	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}

	appG.Success("", nil)
}
func Delete(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, _ := strconv.Atoi(c.Params.ByName("id"))
	err := models.Delete(id)

	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}

	appG.Success("", nil)
}

func UpdateProfile(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var o models.RoastProfile
	if err := c.ShouldBindJSON(&o); err != nil {
		appG.Error(400, "绑定请求中的json到form失败，可能是格式问题", err)
		return
	}
	err := models.UpdateProfile(&o)
	if err != nil {
		appG.Error(500, "更新失败", err)
		return
	}
	appG.Success("", o)
}
func GetProfileById(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, _ := strconv.Atoi(c.Param("id"))
	o, err := models.GetProfileById(id)

	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	appG.Success("", o)
}

func GetProfileByRoastId(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, _ := strconv.Atoi(c.Param("id"))
	o, _ := models.GetProfileByRoastId(uint(id))
	appG.Success("", o)
}
