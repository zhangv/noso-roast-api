package supplier

import (
	"hahacking/common/app"
	"hahacking/middleware"
	models "hahacking/models/supplier"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	APIs := r.Group("/supplier")
	{
		APIs.GET("meta/:ids", Meta)

		userAPI := APIs.Group("")
		userAPI.Use(middleware.JWTAuth())
		userAPI.Use(middleware.BaseAuth("user"))
		{
			userAPI.GET("", Search)
			userAPI.GET(":id", GetById)
			userAPI.POST("", Create)
		}
		adminAPI := APIs.Group("")
		adminAPI.Use(middleware.JWTAuth())
		adminAPI.Use(middleware.BaseAuth("admin"))
		{
			adminAPI.DELETE(":id", Delete)
		}
	}
}

func Meta(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	name := c.Params.ByName("ids")
	result := models.Meta(name)
	appG.Success("", result)
}

type CreateForm struct {
	Name   string `json:"name" form:"name"`
	Code   string `json:"code" form:"code"`
	Type   string `json:"type" form:"type"`
	Status string `json:"status" form:"status"`
	ItemID string `json:"item_id" form:"item_id"`
}
type UpdateForm struct {
	CreateForm
	ID int `json:"id" form:"id" uri:"id" binding:"required"`
}

func Search(c *gin.Context) {
	var f models.SearchForm
	appG := app.Gin{Cxt: c}

	if c.ShouldBindQuery(&f) == nil {
		log.Println(f)
	}

	result, count := models.Search(f)
	appG.Success("", map[string]any{
		"objects": result,
		"total":   count,
	})
}

func Create(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var f CreateForm
	if err := c.ShouldBind(&f); err != nil {
		appG.Failed(400, "参数绑定失败: "+err.Error())
		return
	}

	// 验证必要字段
	if f.Name == "" {
		appG.Failed(400, "name不能为空")
		return
	}

	var o models.Supplier
	o.Name = f.Name
	o.Code = f.Code
	o.Type = models.SupplierType(f.Type)
	o.Status = models.SupplierStatus(f.Status)

	result := models.Update(&o)
	if result != nil {
		appG.Failed(500, "创建供应商失败: "+result.Error())
		return
	}

	appG.Success("供应商创建成功", o)
}
func GetById(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.Atoi(c.Params.ByName("id"))
	result, err := models.GetById(uint(id))

	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}

	appG.Success("", result)
}
func Delete(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	supplier, err := models.GetById(uint(id))
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	err = supplier.Remove()
	if err != nil {
		appG.Failed(500, "删除供应商失败: "+err.Error())
		return
	}

	appG.Success("供应商删除成功", nil)
}
