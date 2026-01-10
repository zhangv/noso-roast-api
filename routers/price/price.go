package price

import (
	"fmt"
	"hahacking/common/app"
	"hahacking/middleware"
	models "hahacking/models/price"
	"hahacking/utils"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

func RegisterRoutes(r *gin.RouterGroup) {
	APIs := r.Group("/price")
	{
	}
	userAPI := APIs.Group("")
	userAPI.Use(middleware.JWTAuth())
	userAPI.Use(middleware.BaseAuth("user"))
	{
		userAPI.GET("", Search)
		userAPI.POST("", Update)
		userAPI.GET(":id", GetById)
		userAPI.GET("level", SearchLevel)
		userAPI.POST("level", UpdateLevel)
		userAPI.GET("level/:id", GetLevelById)

	}
}

type UpdateForm struct {
	ID           uint    `json:"id" form:"id"`
	Price        float64 `json:"price" form:"price"`
	PriceLevelID uint    `json:"pricelevel_id" form:"pricelevel_id"`
}

func (form *UpdateForm) Validate() map[string]string {
	var result = make(map[string]string)
	if form.Price == 0.00 {
		result["price"] = "请填写价格"
	}
	return result
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
	var f models.SearchForm
	appG := app.Gin{Cxt: c}

	if c.ShouldBindQuery(&f) == nil {
		log.Println(f)
	}

	result, count := models.Search(f)
	appG.Success("", utils.Response{
		"objects": result,
		"total":   count,
	})
}

func Update(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var form UpdateForm
	if err := c.ShouldBindJSON(&form); err != nil {
		appG.Error(400, "绑定请求中的json到form失败，可能是格式问题", err)
		return
	}
	err := form.Validate()
	fmt.Errorf("form validation error: %v", err)
	if len(err) > 0 {
		appG.Response(200, 1000, err)
		return
	}

	var o models.Price

	if err := mapstructure.Decode(form, &o); err != nil {
		appG.Error(400, "映射form内容到对象失败", err)
		return
	}
	o.ID = form.ID
	result := models.Update(&o)
	appG.Success("", result)
}
func Delete(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.Atoi(c.Params.ByName("id"))
	err = models.Delete(id)

	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}

	appG.Success("", nil)
}
