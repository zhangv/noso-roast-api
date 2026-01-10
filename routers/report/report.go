package report

import (
	"fmt"
	"hahacking/common/app"
	"hahacking/common/logger"
	"hahacking/common/util"
	"hahacking/middleware"
	models "hahacking/models/report"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func RegisterRoutes(r *gin.RouterGroup) {
	APIs := r.Group("/report")
	{
		// APIs.GET("meta/:ids", Meta)
		userAPI := APIs.Group("")
		userAPI.Use(middleware.JWTAuth())
		userAPI.Use(middleware.BaseAuth("user"))
		{
			userAPI.GET("", Search)
			userAPI.GET("meta/:ids", Meta)
			userAPI.GET(":id", GetById)
			userAPI.GET(":id/status/:to", Status)
			userAPI.GET(":id/columns", GetColumns)
			userAPI.GET(":id/parameters", GetParameters)
			userAPI.GET(":id/sql", GetSQL)
			userAPI.GET(":id/generate", Generate)
			userAPI.POST("", Update)

			userAPI.POST("parameter", UpdateParameter)
			userAPI.DELETE("parameter/:id", DeleteParameter)

			userAPI.POST("column", UpdateColumn)
			userAPI.DELETE("column/:id", DeleteColumn)

			userAPI.GET("chart", SearchChart)
			userAPI.GET("chart/:id", GetChartById)
			userAPI.GET("chart/:id/status/:to", StatusChart)
			userAPI.GET("chart/:id/data", GetChartData)
			userAPI.POST("chart", UpdateChart)

		}
		adminAPI := APIs.Group("")
		adminAPI.Use(middleware.JWTAuth())
		adminAPI.Use(middleware.BaseAuth("admin"))
		{
			adminAPI.DELETE("/:id", Delete)
			adminAPI.DELETE("chart/:id", DeleteChart)
		}
	}
}

func Update(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var f models.Report
	c.ShouldBind(&f)
	errs := f.ValidateForm()
	if len(errs) > 0 {
		appG.Error(500, "表单报错", errs)
		return
	}
	updatedReport, err := models.Update(&f)
	if err != nil {
		appG.Error(500, "抱歉未找到相关信息", err.Error())
		return
	}
	appG.Success("", updatedReport)
}
func Generate(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	report, err := models.GetById(uint(id))
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}

	reqParams := util.GetQueryParams(c)
	//需要检查参数是否合法
	errs := map[string]any{}
	for _, param := range report.Parameters {
		if reqParams[param.Name] == "" {
			theDefaultValue := param.GetDefaultValue()
			if theDefaultValue != "" {
				reqParams[param.Name] = theDefaultValue
			} else {
				err := fmt.Sprintf("缺少参数：%s (%s)", param.Label, param.Name)
				errs[param.Name] = err
				return
			}
		}
	}
	if len(errs) > 0 {
		appG.Error(400, "参数错误", errs)
		return
	}
	results, err := report.Generate(reqParams, 1000)
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	if c.Query("export") != "" {
		if c.Query("export") == "xlsx" {
			exportXls(report, results, c)
			return
		} else {
			appG.Failed(400, "不支持的导出格式")
			return
		}

	} else {
		appG.Success("", results)
	}
}
func exportXls(report *models.Report, results map[string]any, c *gin.Context) {
	logger.Debug("exportXls", report.Name+"("+strconv.Itoa(int(report.ID))+")")
	logger.Debug("exportXls cols", results["cols"])
	f := excelize.NewFile()
	defer f.Close()

	// 设置为活动工作表
	f.SetActiveSheet(0)
	cols := results["cols"].([]models.Column)
	activeSheetIndex := f.GetActiveSheetIndex()
	sheetName := f.GetSheetName(activeSheetIndex)

	logger.Debug("exportXls", "列数:", len(cols), "工作表名:", sheetName)

	// 设置表头
	for j, col := range cols {
		cellName := fmt.Sprintf("%c1", 'A'+j)
		err := f.SetCellValue(sheetName, cellName, col.DisplayName)
		if err != nil {
			logger.Error("设置表头失败", cellName, col.DisplayName, err)
		}
	}

	// 设置单元格格式为常规，避免在excel里认为是文本而不能进行函数公式运算
	styleID, _ := f.NewStyle(&excelize.Style{
		NumFmt: 0, // 0 = 常规
	})
	// 检查数据
	if list, ok := results["list"].([]map[string]any); ok {
		logger.Debug("exportXls", "数据行数:", len(list))
		if len(list) > 0 {
			// 写入数据
			for i, row := range list {
				for j, col := range cols {
					cellName := fmt.Sprintf("%c%d", 'A'+j, i+2)
					value := row[col.Name]
					err := f.SetCellValue(sheetName, cellName, value)
					if err != nil {
						logger.Error("设置单元格失败", cellName, value, err)
					}
					f.SetCellStyle(sheetName, cellName, cellName, styleID)
				}
			}
			logger.Debug("exportXls", "写入数据完成", len(list))
		} else {
			logger.Warn("exportXls", "数据列表为空")
		}
	} else {
		logger.Error("exportXls", "无法获取数据列表，results类型:", fmt.Sprintf("%T", results["list"]))
		if results["list"] != nil {
			logger.Debug("exportXls", "list内容:", results["list"])
		}
	}

	// 设置响应头
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

	// 处理中文文件名编码
	filename := fmt.Sprintf("%s.xlsx", report.Name)
	encodedFilename := url.QueryEscape(filename)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", encodedFilename, encodedFilename))
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	// 写入响应
	if err := f.Write(c.Writer); err != nil {
		appG := app.Gin{Cxt: c}
		appG.Error(500, "导出失败", err.Error())
		logger.Error("导出失败", err.Error())
	} else {
		logger.Debug("导出成功")
	}
}
func GetSQL(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	report, err := models.GetById(uint(id))
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	reqParams := util.GetQueryParams(c)
	sql := report.GenerateSql(reqParams)
	appG.Success("", sql)
}

func GetParameters(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	report, err := models.GetById(uint(id))
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	appG.Success("", report.GetParameters())
}

func GetColumns(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	report, err := models.GetById(uint(id))
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	appG.Success("", report.GetColumns())
}

func Status(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	to := c.Params.ByName("to")
	err = models.Status(id, to)
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	appG.Success("", nil)
}

type CreateForm struct {
	Name        string `form:"name"`
	Code        string `form:"code"`
	Type        string `form:"type"`
	Status      string `form:"status"`
	Description string `form:"description"`
	ParentID    uint   `form:"parent_id"`
}
type UpdateForm struct {
	CreateForm
	ID int `form:"id" uri:"id" binding:"required"`
}

func Meta(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	name := c.Params.ByName("ids")
	result := models.Meta(name)
	appG.Success("", result)
}

func Search(c *gin.Context) {
	var f models.SearchForm
	appG := app.Gin{Cxt: c}

	if c.ShouldBindQuery(&f) == nil {
		log.Println(f)
	}

	result, count := models.Search(f)
	for i, report := range result {
		fmt.Printf("report %d: ID=%d, UserID=%d , Name=%s\n", i, report.ID, report.UserID, report.Name)
	}
	appG.Success("", map[string]any{
		"objects": result,
		"total":   count,
	})
}

func Create(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var f CreateForm
	c.ShouldBind(&f)
	//todo 直接自动映射的方法
	var o models.Report
	o.Name = f.Name

	updatedReport, err := models.Update(&o)
	if err != nil {
		appG.Error(500, "抱歉未找到相关信息", err.Error())
		return
	}

	appG.Success("", updatedReport)
}
func GetById(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	idParam := c.Params.ByName("id")

	// 检测是否是逗号间隔的多个ID
	if strings.Contains(idParam, ",") {
		// 处理多个ID的情况
		ids := strings.Split(idParam, ",")
		var results []interface{}

		for _, idStr := range ids {
			idStr = strings.TrimSpace(idStr) // 去除空格
			if idStr == "" {
				continue
			}

			result, err := models.GetReport(idStr)
			if err != nil {
				// 记录错误但继续处理其他ID
				log.Printf("获取ID %s 失败: %v", idStr, err)
				continue
			}

			results = append(results, result)
		}

		if len(results) == 0 {
			appG.NotFound("抱歉未找到相关信息")
			return
		}

		appG.Success("", results)
		return
	}

	// 处理单个ID的情况
	result, err := models.GetReport(idParam)
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}

	appG.Success("", result)
}

func Delete(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id := c.Params.ByName("id")
	report, err := models.GetReport(id)
	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	err = report.Delete()

	if err != nil {
		appG.NotFound("删除失败")
		return
	}

	appG.Success("", nil)
}
