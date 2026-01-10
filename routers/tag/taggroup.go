package tag

import (
	"hahacking/common/app"
	models "hahacking/models/tag"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

func AddGroup(c *gin.Context) {
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
func GetGroupById(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.Atoi(c.Params.ByName("id"))
	o, err := models.GetGroupById(uint(id))

	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	appG.Success("", o)
}

func SearchGroup(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var f models.SearchGroupForm
	if c.ShouldBindQuery(&f) == nil {
		log.Println(f)
	}
	result := models.SearchGroup(f)
	appG.Success("", result)
}
