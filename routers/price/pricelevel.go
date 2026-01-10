package price

import (
	"hahacking/common/app"
	models "hahacking/models/price"
	"hahacking/utils"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

type UpdateLevelForm struct {
	ID          uint    `json:"id" form:"id"`
	Name        string  `json:"name"`
	Group       string  `json:"group"`
	Discount    float64 `json:"discount"`
	Status      string  `json:"status"`
	Type        string  `json:"type"`
	LowerBound  string  `json:"lowerbound"`
	UpperBound  string  `json:"upperbound"`
	Description string  `json:"description" form:"description"`
}

func (form *UpdateLevelForm) Validate() map[string]string {
	var result = make(map[string]string)
	if form.Name == "" {
		result["name"] = "请填写名称"

	}

	return result
}

func GetLevelById(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.Atoi(c.Params.ByName("id"))
	o, err := models.GetById(uint(id))

	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	appG.Success("", o)
}

func SearchLevel(c *gin.Context) {
	var f models.SearchLevelForm
	appG := app.Gin{Cxt: c}

	if c.ShouldBindQuery(&f) == nil {
		log.Println(f)
	}

	result, count := models.SearchLevel(f)
	appG.Success("", utils.Response{
		"objects": result,
		"total":   count,
	})
}

func UpdateLevel(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	// var jsonData map[string]interface{}
	// if err := c.ShouldBindJSON(&jsonData); err != nil {
	// 	appG.Error(400, "绑定请求中的json到form失败，可能是格式问题", err)
	// 	return
	// }
	var form UpdateLevelForm
	if err := c.ShouldBindJSON(&form); err != nil {
		appG.Error(400, "绑定请求中的json到form失败，可能是格式问题", err)
		return
	}
	err := form.Validate()
	// fmt.Errorf("form validation error: %v", err)
	if len(err) > 0 {
		appG.Response(200, 1000, err)
		return
	}

	var o models.PriceLevel

	if err := mapstructure.Decode(form, &o); err != nil {
		appG.Error(400, "映射form内容到对象失败", err)
		return
	}
	o.ID = form.ID
	result := models.UpdateLevel(&o)
	appG.Success("", result)
}
func DeleteLevel(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.Atoi(c.Params.ByName("id"))
	err = models.DeleteLevel(id)

	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}

	appG.Success("", nil)
}
