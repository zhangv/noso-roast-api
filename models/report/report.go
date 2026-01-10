package report

import (
	"hahacking/common/logger"
	"hahacking/database"
	"hahacking/models"
	"hahacking/models/user"
	"hahacking/utils"
	"reflect"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

/*
 */
type Report struct {
	models.Model
	Code        string        `json:"code" gorm:"column:code"`
	Name        string        `json:"name" gorm:"type:varchar(255);column:name"`
	SqlText     string        `json:"sqltext" gorm:"type:text;column:sqltext"`
	Sort        string        `json:"sort" gorm:"column:sort"`
	Type        ReportType    `json:"type" gorm:"type:varchar(255);column:type"`
	Description string        `json:"description" gorm:"type:text;column:description"`
	Tags        string        `json:"tags" gorm:"type:varchar(255);column:tags"`
	UserID      *uint         `json:"user_id" gorm:"column:user_id;null"`
	User        *user.User    `gorm:"foreignKey:UserID"`
	Status      models.Status `json:"status" gorm:"column:status"`
	Columns     []Column      `json:"columns" gorm:"foreignKey:ReportID;references:ID"`
	Parameters  []Parameter   `json:"parameters" gorm:"foreignKey:ReportID;references:ID"`
}
type ReportType string

const (
	ReportTypeTable ReportType = "table"
	ReportTypeShop  ReportType = "shop"
)

var ReportTypes = map[ReportType]string{
	ReportTypeTable: "表格",
	ReportTypeShop:  "门店",
}

func (r *Report) TableName() string {
	return "t_report"
}

func (r *Report) ValidateForm() map[string]string {
	errs := make(map[string]string)
	if r.Name == "" {
		errs["name"] = "名称不能为空"
	}
	return errs
}
func Meta(name string) map[string]any {
	names := strings.Split(name, ",")
	result := make(map[string]any)
	for _, n := range names {
		switch n {
		case "ReportType", "Type":
			result[n] = models.ConvertMapToArray(ReportTypes)
		case "ReportStatus", "Status":
			result[n] = models.ConvertMapToArray(models.CommonStatuses)
		case "ParameterType":
			result[n] = models.ConvertMapToArray(ParameterTypes)
		case "ChartType":
			result[n] = models.ConvertMapToArray(ChartTypes)
		case "ChartStatus":
			result[n] = models.ConvertMapToArray(models.CommonStatuses)
		}
	}
	return result
}
func GetReport(idorcode string) (*Report, error) {
	var o Report
	db := database.Conn
	err := db.Where("id = ? or code = ?", idorcode, idorcode).
		Preload("User").
		Preload("Columns", "status = ?", models.StatusValid).
		Preload("Parameters", "status = ?", models.StatusValid).
		First(&o).Error
	if err != nil || err == gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}
func GetById(id uint) (*Report, error) {
	var o Report
	db := database.GetConn()
	err := db.Where("id = ?", id).
		Preload("User").
		Preload("Columns", "status = ?", models.StatusValid).
		Preload("Parameters", "status = ?", models.StatusValid).
		First(&o).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

type SearchForm struct {
	utils.Pager
	Name   string `form:"name"`
	Code   string
	Status string `form:"status"`
}

func Search(q SearchForm) ([]Report, int) {
	db := database.GetConn()
	var os []Report
	var statement = db.Model(&Report{}).Preload("User")
	if q.Name != "" {
		statement = statement.Where("name LIKE ?", "%"+q.Name+"%")
	}
	if q.Code != "" {
		statement = statement.Where("code = ?", q.Code)
	}
	if q.Status != "" {
		statement = statement.Where("status = ?", q.Status)
	} else {
		statement = statement.Where("status = ?", models.StatusValid)
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

func Update(o *Report) (*Report, error) {
	db := database.GetConn()
	tx := db.Begin()

	if o.ID != 0 {
		if err := tx.Model(&Report{}).Where("id = ?", o.ID).Updates(&o).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	} else {
		if err := tx.Create(&o).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	logger.Debug("***********", o.ID)
	tx.Commit()
	return o, nil
}

func (r *Report) GetParameters() []Parameter {
	db := database.GetConn()
	var parameters []Parameter
	db.Model(&Parameter{}).Where("report_id = ? and status = ?", r.ID, models.StatusValid).
		Order("sort asc").Find(&parameters)
	return parameters
}
func (r *Report) GetColumns() []Column {
	db := database.GetConn()
	var columns []Column
	db.Model(&Column{}).Where("report_id = ? and status = ?", r.ID, models.StatusValid).
		Order("sort asc").Find(&columns)
	return columns
}
func (r *Report) GenerateSql(reqParams map[string]any) string {
	//处理参数默认值
	reportParams := r.GetParameters()
	for _, reportParam := range reportParams {
		if reqParams[reportParam.Name] == nil {
			if reportParam.DefaultValue != "" {
				reqParams[reportParam.Name] = reportParam.GetDefaultValue()
			}
		} else {
			//处理宏参数
			paramValue := reportParam.Macro(reqParams[reportParam.Name].(string), "")
			reqParams[reportParam.Name] = paramValue
		}
	}
	sql := r.SqlText
	for key, value := range reqParams {
		key = "#" + key + "#"
		sql = strings.ReplaceAll(sql, key, value.(string))
	}
	return sql
}

func (r *Report) Generate(params map[string]any, maxsize int) (map[string]any, error) {
	sql := r.GenerateSql(params)
	db := database.GetConn()
	var count int64
	countSQL := "SELECT COUNT(1) FROM (" + sql + ") AS _t"
	if err := db.Raw(countSQL).Scan(&count).Error; err != nil {
		return nil, err
	}

	logger.Debug(sql)
	rows, err := db.Raw(sql).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	dbcolumns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	results := make([]map[string]any, 0)
	for rows.Next() {
		row := make(map[string]any)
		values := make([]any, len(dbcolumns))
		for i := range values {
			values[i] = new(any)
		}
		err = rows.Scan(values...)
		if err != nil {
			return nil, err
		}
		for i, col := range dbcolumns {
			// 解包指针值
			row[col] = unwrapValue(values[i])
		}
		results = append(results, row)
		if maxsize > 0 && len(results) >= maxsize {
			break
		}
	}
	columns := r.GetColumns() //如果没有配置列，则直接用db的列名
	if len(columns) == 0 {
		for _, col := range dbcolumns {
			columns = append(columns, Column{
				ReportID:    r.ID,
				Name:        col,
				DisplayName: col,
				Sort:        "1",
				Status:      models.StatusValid,
			})
		}
	}

	all := map[string]any{
		"list": results,
		"cnt":  count,
		"sql":  sql,
		"cols": columns,
	}
	return all, nil
}
func unwrapValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		v = val.Elem().Interface()
	}

	// 处理 []byte 类型
	if bytes, ok := v.([]byte); ok {
		// 尝试转换为数字或字符串
		str := string(bytes)
		// 如果是纯数字，尝试转换为 int
		if len(str) > 0 && str[0] >= '0' && str[0] <= '9' {
			// 检查是否全是数字
			allDigits := true
			for _, b := range bytes {
				if b < '0' || b > '9' {
					allDigits = false
					break
				}
			}
			if allDigits {
				// 尝试转换为 int64
				if i, err := strconv.ParseInt(str, 10, 64); err == nil {
					return i
				}
			}
		}
		return str
	}

	return v
}

func Status(id int, to string) error {
	db := database.GetConn()
	tx := db.Begin()

	if err := tx.Model(&Report{}).Where("id = ?", id).Update("status", to).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (r *Report) Delete() error {
	db := database.GetConn()
	tx := db.Begin()

	// Delete the cupping record
	if err := tx.Delete(&Report{}, r.ID).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
