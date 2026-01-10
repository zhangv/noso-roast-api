package cupping

import (
	"hahacking/common/app"
	"hahacking/middleware"
	commonmodels "hahacking/models"
	models "hahacking/models/cupping"
	"hahacking/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListScores lists cupping scores
func SearchScore(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var form models.SearchScoreForm
	if err := c.ShouldBindQuery(&form); err != nil {
		appG.Error(400, "参数错误", err)
		return
	}

	data, total := models.SearchScore(form)
	appG.Success("", utils.Response{
		"objects": data,
		"total":   total,
	})
}

// GetScore gets a score by ID
func GetScoreById(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		appG.Error(400, "无效的ID", err)
		return
	}

	data, err := models.GetScoreById(uint(id))
	if err != nil {
		appG.Error(500, "获取评分失败", err)
		return
	}

	appG.Success("", data)
}

// GetScoresBySample gets all scores for a sample
func GetScoresBySample(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		appG.Error(400, "无效的ID", err)
		return
	}

	data, err := models.GetScoresBySampleId(uint(id))
	if err != nil {
		appG.Error(500, "获取样本评分失败", err)
		return
	}

	appG.Success("", data)
}

type CreateScoreForm struct {
	models.CuppingScore
}

// 在杯测页面中的smaple上添加评分
func CreateScore(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var score models.CuppingScore
	if err := c.ShouldBindJSON(&score); err != nil {
		appG.Error(400, "参数错误", err)
		return
	}

	if err := models.UpdateScore(&score); err != nil {
		appG.Error(500, "创建评分失败", err)
		return
	}

	appG.Success("", score)
}

type CreateScoreByUserForm struct {
	//评分项
	models.CuppingScoreItem
	//cupping信息
	Location string `json:"location"`
	Method   string `json:"method"` //制作方法
	//sample信息
	UserID    uint    `json:"user_id"`
	BeanID    uint    `json:"bean_id"`
	BeanQty   uint    `json:"bean_qty"`
	GrindSize string  `json:"grind_size"`
	WaterTemp float64 `json:"water_temp"`
}

// 用户直接根据豆子添加（自动创建cupping、sample、score）
func CreateScoreByUser(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var form CreateScoreByUserForm
	if err := c.ShouldBindJSON(&form); err != nil {
		appG.Error(400, "参数错误", err)
		return
	}
	_, cuid, ok := middleware.GetCurrentUser(c)
	if !ok {
		appG.Error(401, "请先登录", nil)
		return
	}
	cupping := models.Cupping{
		Location: form.Location,
		Method:   models.CuppingMethod(form.Method),
		UserID:   &cuid,
		Status:   models.CuppingStatusValid,
	}
	models.Update(&cupping)
	cuppingSample := models.CuppingSample{
		CuppingID: cupping.ID,
		BeanID:    form.BeanID,
		BeanQty:   form.BeanQty,
		UserID:    &cuid,
		GrindSize: form.GrindSize,
		WaterTemp: form.WaterTemp,
		Method:    models.CuppingMethod(form.Method),
	}
	models.UpdateSample(&cuppingSample)

	score := models.CuppingScore{
		CuppingID:        cupping.ID,
		CuppingSampleID:  cuppingSample.ID,
		UserID:           cuid,
		BeanID:           &form.BeanID,
		CuppingScoreItem: form.CuppingScoreItem,
		Status:           commonmodels.StatusValid,
	}
	models.UpdateScore(&score)
	appG.Success("", score)
}

// UpdateScore updates an existing score
func UpdateScore(c *gin.Context) {
	appG := app.Gin{Cxt: c}

	var score models.CuppingScore
	if err := c.ShouldBindJSON(&score); err != nil {
		appG.Error(400, "参数错误", err)
		return
	}
	if score.Status == "" {
		score.Status = commonmodels.StatusValid
	}

	if err := models.UpdateScore(&score); err != nil {
		appG.Error(500, "更新评分失败", err)
		return
	}

	appG.Success("", score)
}

// DeleteScore deletes a score by ID
func DeleteScore(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		appG.Error(400, "无效的ID", err)
		return
	}

	if err := models.DeleteScore(uint(id)); err != nil {
		appG.Error(500, "删除评分失败", err)
		return
	}

	appG.Success("", nil)
}

// GetSampleScoreSummary gets the score summary for a sample
func GetSampleScoreSummary(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		appG.Error(400, "无效的ID", err)
		return
	}

	data, err := models.GetSampleScoreSummary(uint(id))
	if err != nil {
		appG.Error(500, "获取样本评分汇总失败", err)
		return
	}

	appG.Success("", data)
}

func GetBeanScoreSummary(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	beanid, err := strconv.ParseUint(c.Param("beanid"), 10, 32)
	if err != nil {
		appG.Error(400, "无效的ID", err)
		return
	}
	data, err := models.GetBeanScoreSummary(uint(beanid))
	if err != nil {
		appG.Error(500, "获取豆子评分汇总失败", err)
		return
	}

	appG.Success("", data)
}
