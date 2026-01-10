package tag

import (
	"hahacking/common/app"
	"hahacking/middleware"
	models "hahacking/models/tag"
	"hahacking/utils"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	APIs := r.Group("/tag")
	{
	}
	userAPI := APIs.Group("")
	userAPI.Use(middleware.JWTAuth())
	userAPI.Use(middleware.BaseAuth("user"))
	{
		userAPI.POST("add", Add)
		userAPI.GET("remove", Remove)                 //从某个target删除标签
		userAPI.GET("getGroups", GetGroups)           //根据target获取全部相关标签组
		userAPI.GET("getTags", GetTags)               //根据target获取全部相关标签
		userAPI.GET("search", Search)                 //比如某个target的全部标签
		userAPI.GET("searchGroup", SearchGroup)       //搜索所有group
		userAPI.GET("getGroupById/:id", GetGroupById) //搜索所有group
	}
}

func Add(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var form models.AddForm
	if err := c.ShouldBind(&form); err != nil {
		appG.Error(400, "绑定请求中的json到form失败，可能是格式问题", err)
		return
	}
	formerror := form.Validate()
	if len(formerror) > 0 {
		appG.Response(200, 1000, formerror)
		return
	}
	rr, err := models.Add(form)
	if err != nil {
		appG.Error(400, "创建标签失败", err)
		return
	}
	appG.Success("", rr)
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

func GetGroups(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, _ := c.GetQuery("id")
	typ, _ := c.GetQuery("type")
	groups, err := models.GetGroupByTarget(id, typ)
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	appG.Success("", groups)
}
func GetTags(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, _ := c.GetQuery("id")
	typ, _ := c.GetQuery("type")
	tags, err := models.GetByTarget(id, typ)

	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	appG.Success("", tags)
}

func Remove(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, _ := c.GetQuery("target_id")
	typ, _ := c.GetQuery("target_type")
	tagid, _ := c.GetQuery("tag_id")
	err := models.RemoveTag(tagid, id, typ)
	if err != nil {
		appG.NotFound("删除标签失败")
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
