package baidu

import (
	"hahacking/common/app"
	"hahacking/utils/baidu"

	"github.com/gin-gonic/gin"
)

func Text2Audio(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	text := c.Param("text")
	audio, err := baidu.Text2Audio(text)
	if err != nil {
		appG.Error(400, err.Error(), nil)
		return
	}
	c.Data(200, "audio/mp3", audio)
}

func Recognize(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	image := c.PostForm("image")
	// logger.Sugar.Info("image", image)
	recognizeResp, err := baidu.Recognize(image)
	if err != nil {
		appG.Error(400, err.Error(), nil)
		return
	}
	appG.Success("", recognizeResp)
}
