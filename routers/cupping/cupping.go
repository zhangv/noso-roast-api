package cupping

import (
	"hahacking/common/app"
	"hahacking/middleware"
	models "hahacking/models/cupping"
	"hahacking/utils"
	"hahacking/utils/api"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	APIs := r.Group("/cupping")
	{
		APIs.GET("meta/:ids", Meta)
	}
	userAPI := APIs.Group("")
	userAPI.Use(middleware.JWTAuth())
	userAPI.Use(middleware.BaseAuth("user"))
	{
		userAPI.GET("", Search)
		userAPI.GET(":id", GetById)
		userAPI.POST("", Update)
		userAPI.DELETE(":id", Delete)

		userAPI.GET("sample", SearchSample)
		userAPI.GET("sample/:id", GetSampleById)
		userAPI.POST("sample", UpdateSample)
		userAPI.GET("sample/:id/scoresummary", GetSampleScoreSummary)

		userAPI.POST("score", UpdateScore)
		userAPI.GET("score", SearchScore)
		userAPI.GET("score/bean/:beanid", GetBeanScoreSummary) // 必须在 score/:id 之前注册，避免路由冲突
		userAPI.GET("score/:id", GetScoreById)
		userAPI.POST("score/useradd", CreateScoreByUser) //用户添加杯测评分 - 仅包含bean和评分项，不包含cupping的相关信息
	}
}

// Meta returns metadata for cupping
func Meta(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	name := c.Params.ByName("ids")
	result := models.Meta(name)
	appG.Success("", result)
}

// GetById returns a cupping by ID
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

// Search searches for cuppings based on criteria
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

type UpdateForm struct {
	ID          uint                   `json:"id"`
	Location    string                 `json:"location"`
	Status      string                 `json:"status"`
	Method      string                 `json:"method"`
	Description string                 `json:"description"`
	Samples     []models.CuppingSample `json:"samples" `
}

func (form *UpdateForm) Validate() map[string]string {
	var result = make(map[string]string)
	log.Println("form", form)
	if form.Location == "" {
		result["location"] = "请填写地点"
	}
	if form.Status == "" {
		result["status"] = "请填写状态"
	}
	if form.Method == "" {
		result["method"] = "请填写杯测方法"
	}
	if len(form.Samples) == 0 {
		result["samples"] = "请填写杯测样本"
	}
	return result
}
func Update(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var form UpdateForm
	var o models.Cupping
	formErrors := api.Bind(&form, &o, c)
	// log.Println("form errors", formErrors)
	// log.Println("form", form.ID, form)
	// log.Println("model", o.ID, o)
	o.ID = form.ID //注意：需要手动设置一下，因为model中的id是嵌入的models.Model，在绑定的时候不会自动设置
	if len(formErrors) > 0 {
		appG.Error(400, "表单验证失败", formErrors)
		return
	}

	err := models.Update(&o)
	if err != nil {
		appG.Error(500, "更新失败", err)
		return
	}

	appG.Success("", o)
}

// Delete deletes a cupping by ID
func Delete(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err != nil {
		appG.Error(400, "无效的ID", err)
		return
	}

	err = models.Delete((id))
	if err != nil {
		appG.Error(500, "删除失败", err)
		return
	}

	appG.Success("", nil)
}
