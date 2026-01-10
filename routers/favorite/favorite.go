package favorite

import (
	"hahacking/common/app"
	"hahacking/middleware"
	rootmodels "hahacking/models"
	models "hahacking/models/favorite"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	APIs := r.Group("/favorite")
	{
		// APIs.GET("meta/:ids", Meta)

		userAPI := APIs.Group("")
		userAPI.Use(middleware.JWTAuth())
		userAPI.Use(middleware.BaseAuth("user"))
		{
			userAPI.GET("", Search)
			userAPI.GET(":id", GetById)
			userAPI.POST("", Create)

			userAPI.GET("add/:item_id/:type", UserAdd)
			userAPI.GET("remove/:item_id/:type", UserRemove)
		}
		adminAPI := APIs.Group("")
		adminAPI.Use(middleware.JWTAuth())
		adminAPI.Use(middleware.BaseAuth("admin"))
		{
			adminAPI.DELETE(":id", Delete)
		}
	}
}

func UserAdd(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	typ := c.Params.ByName("type")
	itemID := c.Params.ByName("item_id")

	// 使用GetCurrentUser函数获取用户信息
	username, uid, ok := middleware.GetCurrentUser(c)
	if !ok {
		appG.Error(401, "无法获取用户信息", nil)
		return
	}

	log.Printf("用户操作: username=%s, uid=%d", username, uid)

	f := models.UserAdd(typ, itemID, uid)
	appG.Success("", f)
}
func UserRemove(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	typ := c.Params.ByName("type")
	itemID := c.Params.ByName("item_id")

	// 使用GetCurrentUser函数获取用户信息
	username, uid, ok := middleware.GetCurrentUser(c)
	if !ok {
		appG.Error(401, "无法获取用户信息", nil)
		return
	}

	log.Printf("用户操作: username=%s, uid=%d", username, uid)

	f := models.UserRemove(typ, itemID, uid)
	appG.Success("", f)
}

// func Meta(c *gin.Context) {
// 	appG := app.Gin{Cxt: c}
// 	name := c.Params.ByName("ids")
// 	result := models.Meta(name)
// 	appG.Success("", result)
// }

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

	var o models.Favorite
	o.ItemID = f.ItemID
	o.Type = f.Type
	o.Status = rootmodels.Status(f.Status)

	// 设置状态
	if f.Status == "" {
		o.Status = rootmodels.StatusValid
	} else {
		o.Status = rootmodels.Status(f.Status)
	}

	// 处理时间字段

	result := models.Update(&o)
	if result != nil {
		appG.Failed(500, "创建banner失败: "+result.Error())
		return
	}

	appG.Success("banner创建成功", o)
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
	err = models.Delete(id)

	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}

	appG.Success("", nil)
}
