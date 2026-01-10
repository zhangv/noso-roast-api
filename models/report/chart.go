package report

import (
	"encoding/json"
	"fmt"
	"hahacking/common/logger"
	"hahacking/database"
	"hahacking/database/datatypes"
	"hahacking/models"
	"hahacking/utils"
	"strconv"

	"gorm.io/gorm"
)

/*
 */
type Chart struct {
	models.Model
	Code        string         `json:"code"`
	Name        string         `json:"name"`
	Type        ChartType      `json:"type"`
	Description string         `json:"description"`
	Tags        *uint          `json:"tags" `
	Size        *int           `json:"size"`
	Params      datatypes.JSON `json:"params" gorm:"type:varchar(2048);column:params"`
	ColNames    datatypes.JSON `json:"colnames" gorm:"type:varchar(2048);column:colnames"`
	ReportID    *uint          `json:"report_id"`
	Report      *Report        `json:"report" gorm:"foreignKey:ReportID;references:ID"`
	Status      models.Status
}

func (r *Chart) TableName() string {
	return "t_reportchart"
}

type ChartType string

const (
	ChartTypeBar     ChartType = "bar"
	ChartTypeLine    ChartType = "line"
	ChartTypePie     ChartType = "pie"
	ChartTypeTable   ChartType = "table"
	ChartTypeMap     ChartType = "map"
	ChartTypeScatter ChartType = "scatter"
	ChartTypeRadar   ChartType = "radar"
	ChartTypeCard    ChartType = "card"
	//以下later
	ChartTypeGauge     ChartType = "gauge"
	ChartTypeWordCloud ChartType = "wordcloud"
	ChartTypeHeatmap   ChartType = "heatmap"
)

var ChartTypes = map[ChartType]string{
	ChartTypeBar:       "柱状图",
	ChartTypeLine:      "折线图",
	ChartTypePie:       "饼图",
	ChartTypeTable:     "表格",
	ChartTypeMap:       "地图",
	ChartTypeScatter:   "散点图",
	ChartTypeRadar:     "雷达图",
	ChartTypeCard:      "卡片",
	ChartTypeGauge:     "仪表盘",
	ChartTypeWordCloud: "词云",
	ChartTypeHeatmap:   "热力图",
}

func (rc *Chart) Plot(reqParams map[string]any) map[string]any {
	//处理报表静态参数
	params := map[string]any{}
	json.Unmarshal(rc.Params, &params)
	//处理请求中包含的动态参数
	for key, value := range reqParams {
		params[key] = value
	}
	report := rc.GetReport()
	size := 1000
	if rc.Size != nil {
		size = *rc.Size
	}
	logger.Debug("********** plot params", params)
	results, err := report.Generate(params, size)
	if err != nil {
		return nil
	}

	cols := report.GetColumns()
	if len(cols) == 0 {
		for _, col := range results["cols"].([]string) {
			cols = append(cols, Column{
				Name:        col,
				DisplayName: col,
				Sort:        "1",
				Status:      models.StatusValid,
			})
		}
	}
	var rcCols []string
	if rc.ColNames != nil {
		json.Unmarshal(rc.ColNames, &rcCols)
	}
	trimedCols := make([]Column, 0)
	for _, col := range cols {
		if containsString(rcCols, col.Name) {
			trimedCols = append(trimedCols, col)
		}
	}

	switch rc.Type {
	case ChartTypeBar:
		return buildBar(trimedCols, results)
	case ChartTypeLine:
		return buildBar(trimedCols, results)
	case ChartTypePie:
		return buildPie(trimedCols, results)
	case ChartTypeTable:
		return buildTable(results)
	// case ChartTypeMap:
	// 	return buildMap(results)
	// case ChartTypeScatter:
	// 	return buildScatter(results)
	case ChartTypeRadar:
		return buildRadar(trimedCols, results)
	case ChartTypeCard:
		return buildCard(results)
		// case ChartTypeGauge:
		// 	return buildGauge(results)
		// case ChartTypeWordCloud:
		// 	return buildWordCloud(results)
		// case ChartTypeHeatmap:
		// 	return buildHeatmap(results)
	}

	return nil
}
func containsString(list []string, target string) bool {
	for _, s := range list {
		if s == target {
			return true
		}
	}
	return false
}

// convertToString 将任意类型转换为字符串
func convertToString(value any) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case int64:
		return strconv.FormatInt(v, 10)
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func buildCard(results map[string]any) map[string]any {
	return buildTable(results)
}

/*
	 雷达图
		每一行是一个维度，
		第一列是维度名称 - 如日期
*/
func buildRadar(cols []Column, results map[string]any) map[string]any {
	var data = make([]map[string]any, 0)
	var dataKeys = make([]string, 0)
	var dataLabels = make(map[string]string)
	categoryCol := cols[0]
	categoryKey := categoryCol.Name
	valueCols := cols[1:]

	for _, valueCol := range valueCols {
		dataKeys = append(dataKeys, valueCol.Name)
		dataLabels[valueCol.Name] = valueCol.DisplayName
	}

	categories := make([]string, 0)       //维度名称数组
	maxValues := make(map[string]float64) //每个维度最大值
	var maxValue, minValue float64        //所有维度的最大值和最小值
	result := results["list"].([]map[string]any)
	for _, row := range result {
		categoryName := convertToString(row[categoryKey])
		categories = append(categories, categoryName)
		values := make([]string, 0)
		for _, dataKey := range dataKeys {
			dbvalue := row[dataKey]
			dataLabel := dataLabels[dataKey]
			value := convertToFloat64(dbvalue)
			if value > maxValues[dataLabel] {
				maxValues[dataLabel] = value
			}
			if value < minValue {
				minValue = value
			}
			if value > maxValue {
				maxValue = value
			}
			values = append(values, strconv.FormatFloat(value, 'f', -1, 64))
		}
		data = append(data, map[string]any{
			"value": values,
			"name":  categoryName,
		})
	}
	var indicator = make([]any, 0)
	for maxName, maxValue := range maxValues {
		indicator = append(indicator, map[string]any{
			"name": maxName,
			"max":  strconv.FormatFloat(maxValue, 'f', -1, 64),
		})
	}
	return map[string]any{
		"legend":    categories,
		"indicator": indicator,
		"data":      data,
	}
}
func convertToFloat64(value any) float64 {
	if value == nil {
		return 0.0
	}
	tmp, _ := strconv.ParseFloat(value.(string), 64)
	return tmp
}

func buildTable(results map[string]any) map[string]any {
	return map[string]any{
		"data": results["list"],
		"cols": results["cols"],
	}
}

// 饼图 只支持一行结果集
func buildPie(cols []Column, results map[string]any) map[string]any {
	result := results["list"].([]map[string]any)
	var data = make([]map[string]any, 0)
	if len(result) > 0 {
		dataRow := result[0]
		for _, col := range cols {
			colname := col.DisplayName
			if colname == "" {
				colname = col.Name
			}
			data = append(data, map[string]any{
				"value": dataRow[col.Name],
				"name":  colname,
			})
		}
	}
	return map[string]any{
		"data": data,
	}
}

func buildBar(cols []Column, results map[string]any) map[string]any {
	categoryCol := cols[0]          // 第一个是分类列(x轴)
	categoryKey := categoryCol.Name //x轴 列名
	valueCols := cols[1:]           // 剩余的是值列(y轴)

	var dataKeys []string
	var dataLabels []string
	var dataLabelMap = make(map[string]string)
	var data = make(map[string]any, 0)

	for _, valueCol := range valueCols {
		dataKeys = append(dataKeys, valueCol.Name)
		dataLabels = append(dataLabels, valueCol.DisplayName)
		dataLabelMap[valueCol.Name] = valueCol.DisplayName
		data[valueCol.DisplayName] = make([]any, 0)
	}

	result := results["list"].([]map[string]any)
	categories := make([]string, 0)
	for _, row := range result {
		categories = append(categories, convertToString(row[categoryKey]))
		for _, dataKey := range dataKeys {
			data[dataLabelMap[dataKey]] = append(data[dataLabelMap[dataKey]].([]any),
				convertToString(row[dataKey]))
		}
	}

	return map[string]any{
		"data":         data,
		"categories":   categories,
		"datalabels":   dataLabels,
		"datalabelmap": dataLabelMap,
		"datakeys":     dataKeys,
		"cols":         cols,
	}
}
func (rc *Chart) GetReport() *Report {
	if rc.Report == nil {
		report, err := GetById(*rc.ReportID)
		if err != nil {
			return nil
		}
		rc.Report = report
	}
	return rc.Report
}
func GetChart(idorcode string) (*Chart, error) {
	var chart *Chart
	if id, err := strconv.Atoi(idorcode); err == nil {
		chart, err = GetChartById(uint(id))
		if err != nil {
			return nil, err
		}
	} else {
		chart, err = GetChartByCode(idorcode)
		if err != nil {
			return nil, err
		}
	}
	return chart, nil
}
func GetChartById(id uint) (*Chart, error) {
	var o Chart
	db := database.Conn
	err := db.Where("id = ?", id).Preload("Report").First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}

type SearchChartForm struct {
	utils.Pager
	Name     string `form:"name"`
	Code     string
	Status   string `form:"status"`
	Type     string `form:"type"`
	ReportID string `form:"report_id"`
}

func SearchChart(q SearchChartForm) ([]Chart, int) {
	db := database.Conn
	var os []Chart
	var statement = db.Model(&Chart{}).Preload("Report")
	if q.Name != "" {
		statement = statement.Where("name LIKE ?", "%"+q.Name+"%")
	}
	if q.Code != "" {
		statement = statement.Where("code = ?", q.Code)
	}
	if q.Type != "" {
		statement = statement.Where("type = ?", q.Type)
	}
	if q.Status != "" {
		statement = statement.Where("status = ?", q.Status)
	} else {
		statement = statement.Where("status = ?", models.StatusValid)
	}
	if q.ReportID != "" {
		statement = statement.Where("report_id = ?", q.ReportID)
	}
	_, pagesize, start, sort := q.Pager.Paging()
	var count int64
	statement.
		Count(&count).
		Order(sort).
		Offset(start).
		Limit(pagesize).
		Find(&os)
	return os, int(count)
}

type UpdateChartForm struct {
	ID          uint          `json:"id"`
	ReportID    uint          `json:"report_id"`
	Name        string        `json:"name"`
	Code        string        `json:"code"`
	Status      models.Status `json:"status"`
	Type        string        `json:"type"`
	Size        string        `json:"size"`
	ParamSet    string        `json:"paramset" gorm:"column:paramset;type:text"`
	Description string        `json:"description" gorm:"column:description;type:text"`
	Params      string        `json:"params" gorm:"column:params;type:text"`
	ColNames    string        `json:"colnames" gorm:"column:colnames;type:text"`
}

func (form *Chart) ValidateForm() map[string]string {
	var result = make(map[string]string)
	if form.Name == "" {
		result["name"] = "请填写名称"
	}
	if form.Code == "" {
		result["code"] = "请填写编码"
	} else {
		exist, err := GetChartByCode(form.Code)
		if err == nil && exist != nil {
			if exist.ID != form.ID {
				result["code"] = "编码已存在"
			}
		} else {
			result["code"] = "编码已存在"
		}
	}
	if form.Status == "" {
		result["status"] = "请选择状态"
	}
	if len(form.ColNames) == 0 {
		result["colnames"] = "至少要选择一列"
	}

	return result
}
func GetChartByCode(code string) (*Chart, error) {
	db := database.Conn
	var o Chart
	err := db.Where("code = ? and status = ?", code, models.StatusValid).Preload("Report").First(&o).Error
	if err != nil || err == gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, err
}

func UpdateChart(o *Chart) error {
	db := database.Conn
	tx := db.Begin()

	if o.ID != 0 {
		// 更新现有记录
		if err := tx.Model(&Chart{}).Where("id = ?", o.ID).Updates(o).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		// 创建新记录
		if err := tx.Create(&o).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()
	return nil
}

func DeleteChart(id int) error {
	db := database.GetConn()
	tx := db.Begin()

	// Delete the cupping record
	if err := tx.Delete(&Chart{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func StatusChart(id int, to string) error {
	db := database.Conn
	if err := db.Model(&Chart{}).Where("id = ?", id).Update("status", to).Error; err != nil {
		return err
	}
	return nil
}
