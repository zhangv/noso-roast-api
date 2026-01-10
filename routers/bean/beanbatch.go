package bean

import (
	"hahacking/common/app"
	"hahacking/database/datatypes"
	models "hahacking/models/bean"
	"hahacking/utils"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

type UpdateBatchForm struct {
	ID         uint                  `json:"id" form:"id"`
	Date       string                `json:"date"`
	TotalPrice float64               `json:"totalprice"`
	TotalQty   int                   `json:"totalqty"`
	Type       models.StockLogType   `json:"type"`
	BeanID     uint                  `json:"bean_id"`
	BeanType   models.BeanType       `json:"bean_type"`
	Receipts   datatypes.StringSlice `json:"receipts"`
}

func (form *UpdateBatchForm) Validate() map[string]string {
	var result = make(map[string]string)
	if form.BeanID == 0 {
		result["bean_id"] = "请选择生豆"
	}
	if form.Date == "" {
		result["date"] = "请填写日期"
	}
	if form.TotalPrice == 0.00 {
		result["totalqty"] = "请填写总价值"
	}
	if form.TotalQty == 0 {
		result["totalqty"] = "请填写总数量"
	}
	if form.BeanType == "" {
		result["bean_type"] = "请选择豆类型"
	}
	if form.Type == "" {
		result["type"] = "请选择入库类型"
	}
	return result
}
func GetBatchById(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, _ := strconv.Atoi(c.Params.ByName("id"))
	o, err := models.GetById(uint(id))

	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	appG.Success("", o)
}

func SearchBatch(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var f models.SearchBatchForm

	if c.ShouldBindQuery(&f) == nil {
		log.Println(f)
	}

	result, count := models.SearchBatch(f)
	appG.Success("", utils.Response{
		"objects": result,
		"total":   count,
	})
}

func SearchBatchLog(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var f models.SearchBatchLogForm

	if c.ShouldBindQuery(&f) == nil {
		log.Println(f)
	}

	result, count := models.SearchBatchLog(f)
	appG.Success("", utils.Response{
		"objects": result,
		"total":   count,
	})
}

func UpdateBatch(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var form UpdateBatchForm

	if err := c.ShouldBind(&form); err != nil {
		appG.Error(400, "绑定请求中的json到form失败，可能是格式问题", err)
		return
	}
	err := form.Validate()
	if len(err) > 0 {
		appG.Response(200, 1000, err)
		return
	}

	var o models.BeanBatch

	if err := mapstructure.Decode(form, &o); err != nil {
		appG.Error(400, "映射form内容到对象失败", err)
		return
	}
	o.ID = form.ID

	batch, err2 := models.UpdateBatch(&o)
	if err2 != nil {
		appG.Error(400, "更新批次失败", err2)
		return
	}
	appG.Success("", batch)
}

func UndoBatch(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, _ := strconv.Atoi(c.Params.ByName("id"))
	o, err := models.GetBatchById(uint(id))

	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	err = o.Undo()
	if err != nil {
		appG.Error(400, "撤销失败", err)
		return
	}
	appG.Success("", o)
}
