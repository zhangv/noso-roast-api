package bean

import (
	"fmt"
	"hahacking/common/app"
	"hahacking/middleware"
	models "hahacking/models/bean"
	"hahacking/models/price"
	"hahacking/models/supplier"
	"hahacking/models/tag"
	"hahacking/utils"
	"log"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

func RegisterRoutes(r *gin.RouterGroup) {
	APIs := r.Group("/bean")
	{
		APIs.GET("meta/:ids", Meta)
		APIs.GET("allprice", GetAllPriceList) // 获取所有价格列表 h5首页
	}
	userAPI := APIs.Group("")
	userAPI.Use(middleware.JWTAuth())
	userAPI.Use(middleware.BaseAuth("user"))
	{
		userAPI.GET("", Search)
		userAPI.GET("total", GetTotal)
		userAPI.GET(":id", GetById)
		userAPI.GET(":id/generateprice", GeneratePrice)
		userAPI.GET("batch", SearchBatch)
		userAPI.GET("batch/:id", GetBatchById)
		userAPI.POST("batch", UpdateBatch)
		userAPI.GET("batch/:id/undo", UndoBatch)
		userAPI.GET("batchlog", SearchBatchLog)
		userAPI.GET("stocklog", SearchStockLog)
		userAPI.POST("stocklog", UpdateStockLog)
		userAPI.GET("pricelist", GetPriceList) //单/多个bean的价格列表
		userAPI.POST("", Update)
		userAPI.DELETE(":id", Remove)
	}

	adminAPI := APIs.Group("")
	adminAPI.Use(middleware.JWTAuth())
	adminAPI.Use(middleware.BaseAuth("admin"))
	{

	}
}

/*
查找bean，并返回价格列表和tag列表
*/
func GetAllPriceList(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var f models.SearchForm
	if c.ShouldBindQuery(&f) == nil {
		log.Println(f)
	}
	result, count := models.Search(f)
	for i, b := range result {
		prices, _ := b.GetPriceList(price.LevelTypeC)
		tags, _ := tag.GetByTarget(strconv.Itoa(int(b.ID)), "bean")
		result[i].Prices = prices
		result[i].Tags = tags
	}
	appG.Success("", utils.Response{
		"objects": result,
		"total":   count,
	})
}

func GetTotal(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	total, err := models.GetTotal()
	if err != nil {
		appG.Error(500, "获取总数量失败", err)
		return
	}
	appG.Success("", total)
}

type UpdateForm struct {
	ID                uint          `json:"id" form:"id"`
	Name              string        `json:"name"`
	Code              string        `json:"code"`
	Status            string        `json:"status"`
	Type              string        `json:"type"`
	Variety           string        `json:"variety"`
	Moisture          float64       `json:"moisture"`
	Density           float64       `json:"density"`
	Mesh              string        `json:"mesh"`
	Process           string        `json:"process"`
	OtherProcess      string        `json:"otherprocess"`
	Country           string        `json:"country"`
	Processor         string        `json:"processor"`
	Provider          string        `json:"provider"`
	PurchasePrice     float64       `json:"purchaseprice"`
	BPrice            float64       `json:"bprice"`
	CPrice            float64       `json:"cprice"`
	Owner             string        `json:"owner"`
	Cover             string        `json:"cover"`
	Description       string        `json:"description"`
	RawStockAlert     int           `json:"rawstockalert"`
	RoastedStockAlert int           `json:"roastedstockalert"`
	Prices            []price.Price `json:"prices" gorm:"-"`
}

func (form *UpdateForm) Validate() map[string]string {
	var result = make(map[string]string)
	if form.Name == "" {
		result["name"] = "请填写名称"
	}
	if form.Code == "" {
		result["code"] = "请填写编码"
	} else {
		existingBean, err := models.GetByCode(form.Code) // Changed from form["code"] to form.Code
		if err == nil && existingBean.ID != 0 && existingBean.ID != form.ID {
			errstr := fmt.Sprintf("编码(%s)已存在 ", existingBean.Code)
			result["code"] = errstr
		}
	}
	// if form.Type == "" {
	// 	result["type"] = "请填写类型"
	// }
	if form.Variety == "" {
		result["variety"] = "请填写品种"
	}
	if form.Country == "" {
		result["country"] = "请填写产地"
	}
	if form.Processor == "" {
		result["processor"] = "请填写供应商"
	}
	if form.Process == "" {
		result["process"] = "请填写处理方式"
	}
	if form.Provider == "" {
		result["provider"] = "请填写供应商"
	}
	if form.PurchasePrice == 0.00 {
		result["purchaseprice"] = "请填写采购价"
	}
	if form.CPrice == 0.00 {
		result["cprice"] = "请填写零售价"
	}
	if form.BPrice == 0.00 {
		result["bprice"] = "请填写批发价"
	}
	if form.Owner == "" {
		result["owner"] = "请填写归属"
	}

	return result
}
func Update(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var form UpdateForm
	if err := c.ShouldBindJSON(&form); err != nil {
		appG.Error(400, "绑定请求中的json到form失败，可能是格式问题", err)
		return
	}
	err := form.Validate()
	if len(err) > 0 {
		appG.Response(200, 1000, err)
		return
	}

	var o models.Bean

	if err := mapstructure.Decode(form, &o); err != nil {
		appG.Error(400, "映射form内容到对象失败", err)
		return
	}
	o.ID = form.ID
	result, _ := models.Update(&o)
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

func Remove(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.Atoi(c.Params.ByName("id"))
	bean, err := models.GetById(uint(id))
	err = bean.Remove()
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
func GeneratePrice(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.Atoi(c.Params.ByName("id"))
	overwriteP := c.Query("overwrite")
	var overwrite bool
	if overwriteP == "" {
		overwrite = false
	} else {
		overwrite = true
	}

	if err != nil {
		appG.Error(400, "参数错误", err)
		return
	}
	bean, _ := models.GetById(uint(id))
	result := bean.GeneratePrices(overwrite)
	appG.Success("", result)
}

func GetPriceList(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	beanIDs := c.Query("beanids")
	if beanIDs != "" {
		appG.Error(400, "缺少beanid", nil)
		return
	}
	leveltype := c.Query("leveltype")
	idarray := strings.Split(beanIDs, ",")
	var result []map[string]any
	for _, idStr := range idarray {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			appG.Error(400, "无效的beanid", err)
			return
		}
		bean, _ := models.GetById(uint(id))
		o, err := bean.GetPriceList(price.LevelType(leveltype))
		if err != nil {
			appG.Error(400, "获取价格列表失败", err)
			return
		}

		var tmp = map[string]any{"bean": bean, "prices": o}
		result = append(result, tmp)
	}
	appG.Success("", result)
}

func Meta(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	name := c.Params.ByName("ids")
	result := models.Meta(name)
	appG.Success("", result)
}

func Delete(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, _ := strconv.Atoi(c.Params.ByName("id"))
	bean, _ := models.GetById(uint(id))
	err := bean.Delete()
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}

	appG.Success("", nil)
}

func SearchSupplier(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var form supplier.SearchForm
	if err := c.ShouldBindQuery(&form); err != nil {
		appG.Error(400, "绑定请求中的json到form失败，可能是格式问题", err)
		return
	}
	os, count := supplier.Search(form)
	appG.Success("", utils.Response{
		"objects": os,
		"total":   count,
	})
}
