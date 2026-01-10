package shop

import (
	"hahacking/common/app"
	"hahacking/common/logger"
	models "hahacking/models/shop"
	"hahacking/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

type UpdateAccountForm struct {
	ID      uint    `json:"id" form:"id"`
	Name    string  `json:"name"`
	Type    string  `json:"type"`
	Code    string  `json:"code"`
	Balance float64 `json:"balance" gorm:"type:decimal(10,2)"`
	Status  string  `json:"status" `
	ShopID  uint    `json:"shop_id" `
}

func (form *UpdateAccountForm) Validate() map[string]string {
	var result = make(map[string]string)
	if form.Name == "" {
		result["name"] = "请填写名称"
	}
	if form.Code == "" {
		result["code"] = "请填写编码"
	}
	if form.Type == "" {
		result["type"] = "请选择类型"
	}
	if form.Status == "" {
		result["status"] = "请选择状态"
	}

	return result
}

func GetAccountById(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, _ := strconv.Atoi(c.Params.ByName("id"))
	o, err := models.GetAccountById(id)

	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	appG.Success("", o)
}

func SearchAccount(c *gin.Context) {
	var f models.SearchAccountForm
	appG := app.Gin{Cxt: c}

	err := c.ShouldBindQuery(&f)
	if err != nil {
		appG.Error(400, "绑定请求中的query到form失败，可能是格式问题", err)
		return
	}
	logger.Debugf("SearchAccount: %+v", f)

	result, count := models.SearchAccount(f)
	appG.Success("", utils.Response{
		"objects": result,
		"total":   count,
	})
}

func UpdateAccount(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var form UpdateAccountForm
	if err := c.ShouldBind(&form); err != nil {
		appG.Error(400, "绑定请求中的json到form失败，可能是格式问题", err)
		return
	}
	err := form.Validate()
	if len(err) > 0 {
		appG.Response(200, 1000, err)
		return
	}

	var o models.ShopAccount

	if err := mapstructure.Decode(form, &o); err != nil {
		appG.Error(400, "映射form内容到对象失败", err)
		return
	}
	o.ID = form.ID
	result, err2 := models.UpdateAccount(&o)
	if err2 != nil {
		appG.Error(400, "更新账户失败", err2)
		return
	}
	appG.Success("", result)
}
func DeleteAccount(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, _ := strconv.Atoi(c.Params.ByName("id"))
	err := models.DeleteAccount(id)

	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}

	appG.Success("", nil)
}
