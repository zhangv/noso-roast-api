package report

import (
	"hahacking/database/datatypes"
	"hahacking/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetChartById(t *testing.T) {
	setup()
	rpt := NewReport()
	c1 := Column{
		ReportID:    rpt.ID,
		Name:        "col1",
		DisplayName: "第一列",
		Sort:        "1",
		Status:      models.StatusValid,
	}
	UpdateColumn(&c1)
	c2 := Column{
		ReportID:    rpt.ID,
		Name:        "col2",
		DisplayName: "第二列",
		Sort:        "2",
		Status:      models.StatusValid,
	}
	UpdateColumn(&c2)

	p1 := Parameter{
		ReportID:     rpt.ID,
		Name:         "p1",
		Label:        "参数1",
		Type:         ParameterTypeSelect,
		DefaultValue: "1",
		Description:  "1: 值1, 2: 值2",
		Status:       models.StatusValid,
	}
	UpdateParameter(&p1)
	p2 := Parameter{
		ReportID: rpt.ID,
		Name:     "p2",
		Label:    "参数2",
		Type:     ParameterTypeText,
		Status:   models.StatusValid,
	}
	UpdateParameter(&p2)

	o := Chart{
		Name:     "test",
		Code:     "test" + time.Now().Format("20060102150405"),
		Status:   models.StatusValid,
		ColNames: datatypes.JSON(`["col1","col2"]`),
		// ParamSet:    datatypes.JSON(`{"key":"p1","label":"参数1"}`),
		Params:      datatypes.JSON(`{"p1":"1"}`),
		Description: "测试",
		ReportID:    &rpt.ID,
	}

	UpdateChart(&o)

	r, _ := GetChartById(o.ID)
	assert.NotNil(t, r)
	assert.Equal(t, o.Name, r.Name)
	assert.Equal(t, o.Code, r.Code)
	assert.Equal(t, o.Status, r.Status)
	assert.Equal(t, o.Description, r.Description)

}
