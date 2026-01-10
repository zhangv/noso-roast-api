package report

import (
	"hahacking/database"
	"hahacking/models"
	"hahacking/utils"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Parameter struct {
	models.Model
	Name         string        `gorm:"column:name" json:"name"`
	Code         string        `gorm:"column:code" json:"code"`
	Label        string        `gorm:"column:label" json:"label"`
	Sort         string        `gorm:"column:sort" json:"sort"`
	Status       models.Status `gorm:"column:status" json:"status"`
	Description  string        `gorm:"type:text;column:description" json:"description"`
	DefaultValue string        `gorm:"type:text;column:defaultvalue" json:"defaultvalue"`
	ReportID     uint          `gorm:"column:report_id" json:"report_id"`
	Report       Report        `gorm:"foreignKey:ReportID;references:ID"`
	Type         ParameterType `json:"type"`
}

func (p *Parameter) TableName() string {
	return "t_reportparameter"
}

type ParameterType string

const (
	ParameterTypeText        ParameterType = "text"
	ParameterTypeDate        ParameterType = "date"
	ParameterTypeTime        ParameterType = "time"
	ParameterTypeSelect      ParameterType = "select"
	ParameterTypeMultiSelect ParameterType = "multiselect"
	ParameterTypeCheckbox    ParameterType = "checkbox"
	ParameterTypeRadio       ParameterType = "radio"
	ParameterTypeDateRange   ParameterType = "daterange"
	ParameterTypeTimeRange   ParameterType = "timerange"
)

var ParameterTypes = map[ParameterType]string{
	ParameterTypeText:        "文本",
	ParameterTypeDate:        "日期",
	ParameterTypeTime:        "时间",
	ParameterTypeSelect:      "单选",
	ParameterTypeMultiSelect: "多选",
	ParameterTypeCheckbox:    "复选框",
	ParameterTypeRadio:       "单选框",
	ParameterTypeDateRange:   "日期范围",
	ParameterTypeTimeRange:   "时间范围",
}

func (p *Parameter) ValidateForm() map[string]string {
	errs := make(map[string]string)
	if p.ReportID == 0 {
		errs["report_id"] = "报表ID不能为空"
	}
	if p.Name == "" {
		errs["name"] = "名称不能为空"
	}
	return errs
}
func (p *Parameter) GetDefaultValue() string {
	return p.Macro(p.DefaultValue, "").(string)
}
func (p *Parameter) Macro(macro string, defaultValue interface{}) interface{} {
	switch macro {
	case "timestamp()":
		return time.Now().Unix()
	case "time()":
		return time.Now().Format("2006-01-02 15:04:05")
	case "date()":
		return time.Now().Format("2006-01-02")
	case "datetime()":
		return time.Now().Format("2006-01-02 15:04:05")
	case "earliest()":
		return int64(0)
	case "earliesttime()":
		return time.Unix(0, 0).Format("2006-01-02 15:04:05")
	case "earliestdate()":
		return time.Unix(0, 0).Format("2006-01-02")
	case "CURRENTMONTH_D1":
		now := time.Now()
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	case "CURRENTMONTH":
		return time.Now().Format("2006-01")
	default:
		// 处理 strtotime() 函数
		if strings.Contains(strings.ToLower(macro), "strtotime") {
			re := regexp.MustCompile(`(?i)strtotime\((.*)\)`)
			matches := re.FindStringSubmatch(macro)
			if len(matches) > 1 {
				// 移除引号和括号
				timeStr := strings.Trim(strings.Trim(matches[1], "'"), "()")
				if t, err := time.Parse("2006-01-02 15:04:05", timeStr); err == nil {
					return t.Format("2006-01-02 15:04:05")
				}
				// 尝试解析相对时间
				if t, err := utils.ParseRelativeTime(timeStr); err == nil {
					return t.Format("2006-01-02 15:04:05")
				}
			}
		}

		// 处理 strtodate() 函数
		if strings.Contains(strings.ToLower(macro), "strtodate") {
			re := regexp.MustCompile(`(?i)strtodate\((.*)\)`)
			matches := re.FindStringSubmatch(macro)
			if len(matches) > 1 {
				// 移除引号和括号
				timeStr := strings.Trim(strings.Trim(matches[1], "'"), "()")
				if t, err := utils.ParseRelativeTime(timeStr); err == nil {
					return t.Format("2006-01-02")
				}
			}
		}

		// 处理 date() 函数
		if strings.Contains(macro, "date(") {
			re := regexp.MustCompile(`date\((.*),(.*)\)`)
			matches := re.FindStringSubmatch(macro)
			if len(matches) > 2 {
				tmp, _ := utils.PHPTimeFormatRelativeFixed(strings.Trim(matches[1], "'"), strings.Trim(matches[2], "'"))
				return tmp
			}
		}

		return macro
	}
}

func GetParameterById(id uint) (*Parameter, error) {
	var o Parameter
	db := database.Conn
	err := db.Where("id = ?", id).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}

type SearchParameterForm struct {
	utils.Pager
	Name     string `form:"name"`
	Code     string
	Status   string `form:"status"`
	ReportID string `form:"report_id"`
}

func SearchParameter(q SearchParameterForm) ([]Parameter, int) {
	db := database.Conn
	var os []Parameter
	var statement = db.Model(&Parameter{})
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

func UpdateParameter(o *Parameter) error {
	db := database.Conn
	tx := db.Begin()

	if o.ID != 0 {
		if err := tx.Model(&Parameter{}).Where("id = ?", o.ID).Updates(&o).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := tx.Create(&o).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()
	return nil
}

func StatusParameter(id int, to string) error {
	db := database.GetConn()
	tx := db.Begin()

	if err := tx.Model(&Parameter{}).Where("id = ?", id).Update("status", to).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func DeleteParameter(id int) error {
	db := database.GetConn()
	tx := db.Begin()

	// Delete the cupping record
	if err := tx.Delete(&Parameter{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
