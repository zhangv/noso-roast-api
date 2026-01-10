package bean

import (
	"hahacking/common/app"
	"hahacking/middleware"
	models "hahacking/models/bean"
	"hahacking/utils"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

func SearchStockLog(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var f models.SearchStockLogForm

	if c.ShouldBindQuery(&f) == nil {
		log.Println(f)
	}
	result, count := models.SearchStockLog(f)
	appG.Success("", utils.Response{
		"objects": result,
		"total":   count,
	})
}

type UpdateStockLogForm struct {
	ID          uint                `json:"id" form:"id"`
	TotalPrice  float64             `json:"totalprice"`
	TotalQty    int                 `json:"totalqty"`
	Type        models.StockLogType `json:"type"`
	BeanID      uint                `json:"bean_id"`
	BeanBatchID uint                `json:"beanbatch_id"`
	BeanType    models.BeanType     `json:"bean_type"`
}

func (form *UpdateStockLogForm) Validate() map[string]string {
	var result = make(map[string]string)
	if form.BeanID == 0 {
		result["bean_id"] = "请选择豆"
	}
	if form.BeanType == "" {
		result["bean_type"] = "请选择豆类型"
	}
	if form.TotalPrice == 0.00 {
		result["totalprice"] = "请填写总价值"
	}
	if form.TotalQty == 0 {
		result["totalqty"] = "请填写总数量"
	}
	if form.Type == "" {
		result["type"] = "请选择库存类型"
	}
	return result
}
func UpdateStockLog(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var form UpdateStockLogForm
	if err := c.ShouldBind(&form); err != nil {
		appG.Error(400, "绑定请求中的json到form失败，可能是格式问题", err)
		return
	}
	err := form.Validate()
	if len(err) > 0 {
		appG.Response(200, 1000, err)
		return
	}

	var o models.BeanStockLog

	if err := mapstructure.Decode(form, &o); err != nil {
		appG.Error(400, "映射form内容到对象失败", err)
		return
	}

	bean, _ := models.GetById(form.BeanID)
	if bean == nil {
		appG.Error(400, "未找到相关生豆信息", nil)
		return
	}
	_, userid, ok := middleware.GetCurrentUser(c)
	if !ok {
		appG.Error(400, "未登录", nil)
		return
	}
	delta := form.TotalQty
	if contains(models.StockOutTypes(), string(form.Type)) {
		delta = -delta
	}
	_, err2 := bean.Stocking(delta, form.BeanType, form.Type, "user", strconv.Itoa(int(userid)),
		"手动修改库存"+strconv.Itoa(int(form.ID)))
	if err2 != nil {
		appG.Error(400, "库存修改失败", err2)
		return
	}
	if form.BeanBatchID > 0 {
		beanbatch, _ := models.GetBatchById(form.BeanBatchID)
		if beanbatch == nil {
			appG.Error(400, "未找到相关批次信息", nil)
			return
		}
		beanbatch.Log(delta, "user", strconv.Itoa(int(userid)),
			"手动修改库存"+strconv.Itoa(int(form.ID)))
	}

	appG.Success("", delta)
}
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
