package shop

import (
	"hahacking/common/app"
	"hahacking/middleware"
	models "hahacking/models/shop"
	"hahacking/utils"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

func RegisterRoutes(r *gin.RouterGroup) {
	APIs := r.Group("/shop")
	{
		APIs.GET("meta/:ids", Meta)
	}
	userAPI := APIs.Group("")
	userAPI.Use(middleware.JWTAuth())
	userAPI.Use(middleware.BaseAuth("user"))
	{
		userAPI.GET("account", SearchAccount)
		userAPI.GET("account/:id", GetAccountById)
		userAPI.POST("account", UpdateAccount)
		userAPI.GET("accountlog", SearchAccountLog)
		userAPI.GET("accountlog/:id", GetAccountLogById)
	}
}

func Meta(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	name := c.Params.ByName("ids")
	result := models.Meta(name)
	appG.Success("", result)
}

type UpdateForm struct {
	ID            uint   `json:"id" form:"id"`
	Name          string `json:"name"`
	Code          string `json:"code"`
	Status        string `json:"status"`
	Type          string `json:"type"`
	Processor     string `json:"processor"`
	Variety       string `json:"variety"`
	Process       string `json:"process"`
	OtherProcess  string `json:"otherprocess"`
	Country       string `json:"country"`
	PurchasePrice string `json:"purchaseprice"`
	BPrice        string `json:"bprice"`
	CPrice        string `json:"cprice"`
	Owner         string `json:"owner"`
}

func (form *UpdateForm) Validate() map[string]string {
	var result = make(map[string]string)
	if form.Name == "" {
		result["name"] = "请填写名称"
	}
	if form.Code == "" {
		result["code"] = "请填写编码"
	} else {

	}

	return result
}

type SearchForm struct {
	utils.Pager
	Name   string `form:"name"`
	Code   string `form:"code"`
	Status string `form:"status"`
}

func GetById(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, _ := strconv.Atoi(c.Params.ByName("id"))
	o, err := models.GetById(id)

	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	appG.Success("", o)
}

func Search(c *gin.Context) {
	var f SearchForm
	appG := app.Gin{Cxt: c}

	if c.ShouldBindQuery(&f) == nil {
		log.Println(f)
	}

	o := map[string]interface{}{}
	if f.Name != "" {
		o["name"] = f.Name
	}
	if f.Status != "" {
		o["status"] = f.Status
	}
	o["pageNo"] = f.PageNo
	o["pageSize"] = f.PageSize
	result, count := models.Search(o)
	appG.Success("", utils.Response{
		"objects": result,
		"total":   count,
	})
}

func Update(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var form UpdateForm
	if err := c.ShouldBind(&form); err != nil {
		appG.Error(400, "绑定请求中的json到form失败，可能是格式问题", err)
		return
	}
	err := form.Validate()
	if len(err) > 0 {
		appG.Response(200, 1000, err)
		return
	}

	var o models.Shop

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
	id, _ := strconv.Atoi(c.Params.ByName("id"))
	err := models.Delete(id)

	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}

	appG.Success("", nil)
}
