package cupping

import (
	"hahacking/common/app"
	models "hahacking/models/cupping"
	"hahacking/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListSamples lists cupping samples
func SearchSample(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var form models.SearchSampleForm
	if err := c.ShouldBindQuery(&form); err != nil {
		appG.Error(400, "参数错误", err)
		return
	}

	data, total := models.SearchSample(form)
	appG.Success("", utils.Response{
		"objects": data,
		"total":   total,
	})
}

// GetSample gets a sample by ID
func GetSampleById(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		appG.Error(400, "无效的ID", err)
		return
	}

	data, err := models.GetSampleById(uint(id))
	if err != nil {
		appG.Error(500, "获取样本失败", err)
		return
	}

	appG.Success("", data)
}

// CreateSample creates a new sample
func CreateSample(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var sample models.CuppingSample
	if err := c.ShouldBindJSON(&sample); err != nil {
		appG.Error(400, "参数错误", err)
		return
	}

	if err := models.UpdateSample(&sample); err != nil {
		appG.Error(500, "创建样本失败", err)
		return
	}

	appG.Success("", sample)
}

// UpdateSample updates an existing sample
func UpdateSample(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var sample models.CuppingSample
	if err := c.ShouldBindJSON(&sample); err != nil {
		appG.Error(400, "参数错误", err)
		return
	}

	if err := models.UpdateSample(&sample); err != nil {
		appG.Error(500, "更新样本失败", err)
		return
	}

	appG.Success("", sample)
}

// DeleteSample deletes a sample by ID
func DeleteSample(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		appG.Error(400, "无效的ID", err)
		return
	}

	sample, err := models.GetSampleById(uint(id))
	if err != nil {
		appG.Error(500, "获取样本失败", err)
		return
	}
	if err := sample.Delete(); err != nil {
		appG.Error(500, "删除样本失败", err)
		return
	}

	appG.Success("", nil)
}
