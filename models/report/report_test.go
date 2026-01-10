package report

import (
	"fmt"
	"hahacking/common/logger"
	"hahacking/config"
	"hahacking/database"
	"hahacking/models"
	"hahacking/models/user"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	gormlogger "gorm.io/gorm/logger"
)

var once sync.Once

func setup() {
	once.Do(func() {
		config.Setup("../../", "test") //相对当前路径
		database.SetLogLevel(gormlogger.Info)
		database.Setup()
		logger.Setup()

		db := database.GetConn()
		db.AutoMigrate(
			&Report{}, &Parameter{}, &Column{}, &Chart{},
			&user.User{},
		)
	})
}

func NewReport() *Report {
	u := user.User{
		Nickname: "testnick" + time.Now().Format("2006010215040500000"),
	}
	user.Update(&u)
	uid := u.ID
	r := Report{
		Name:    "testing",
		SqlText: "select * from t_report where id=#p1#",
		Status:  models.StatusValid,
		UserID:  &uid,
	}
	Update(&r)

	return &r
}

func TestGetById(t *testing.T) {
	setup()
	u := user.User{
		Nickname: "testnick",
		Mobile:   nil,
	}
	user.Update(&u)
	uid := u.ID
	o := Report{
		Name:    "testing",
		SqlText: "select * from t_report where id=#p1#",
		Status:  models.StatusValid,
		UserID:  &uid,
	}

	Update(&o)

	r, _ := GetById(o.ID)
	assert.NotNil(t, r)
	assert.NotNil(t, r.User)
	assert.Equal(t, u.Nickname, r.User.Nickname)

}

func TestGenerate(t *testing.T) {
	setup()

	o := Report{
		Name:    "testing",
		SqlText: "select * from t_report where id=#p1#",
		Status:  models.StatusValid,
	}
	updatedReport, _ := Update(&o)
	parameter := Parameter{
		ReportID:     updatedReport.ID,
		Name:         "p1",
		Label:        "参数1",
		DefaultValue: fmt.Sprintf("%d", o.ID),
		Sort:         "1",
		Status:       models.StatusValid,
	}
	UpdateParameter(&parameter)
	logger.Debug(updatedReport.ID)
	res, _ := updatedReport.Generate(map[string]any{
		"p1": fmt.Sprintf("%d", updatedReport.ID),
	}, 10)
	list := res["list"].([]map[string]any)
	assert.NotNil(t, list[0])
	assert.Equal(t, int64(o.ID), list[0]["id"])

	res, _ = updatedReport.Generate(map[string]any{}, 10)
	list = res["list"].([]map[string]any)
	assert.NotNil(t, list[0])
	assert.Equal(t, int64(o.ID), list[0]["id"])

	err := o.Delete()
	assert.Nil(t, err)
	exists, _ := GetById(o.ID)
	assert.Nil(t, exists)

}

func TestGetByIdWithPreload(t *testing.T) {
	setup()

	// 创建测试用户
	u := user.User{
		Nickname: "testnick" + time.Now().Format("2006010215040500000"),
	}
	user.Update(&u)
	uid := u.ID

	// 创建测试报告
	o := Report{
		Name:    "testing preload",
		SqlText: "select `a` as `col1`,`b` as `col2` from t_report where id=#p1#",
		Status:  models.StatusValid,
		UserID:  &uid,
	}
	Update(&o)

	// 创建测试列
	column := Column{
		ReportID:    o.ID,
		Name:        "test_column",
		DisplayName: "测试列",
		Sort:        "1",
		Status:      models.StatusValid,
	}
	UpdateColumn(&column)

	// 创建测试参数
	parameter := Parameter{
		ReportID:     o.ID,
		Name:         "p1",
		Label:        "参数1",
		DefaultValue: "1",
		Sort:         "1",
		Status:       models.StatusValid,
		Type:         ParameterTypeText,
	}
	UpdateParameter(&parameter)

	// 测试 GetById 是否能正确 Preload 关联数据
	r, err := GetById(o.ID)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	assert.NotNil(t, r.User)
	assert.Equal(t, u.Nickname, r.User.Nickname)

	// 验证 Columns 是否正确加载
	assert.NotNil(t, r.Columns)
	assert.Len(t, r.Columns, 1)
	assert.Equal(t, "test_column", r.Columns[0].Name)
	assert.Equal(t, "测试列", r.Columns[0].DisplayName)

	// 验证 Parameters 是否正确加载
	assert.NotNil(t, r.Parameters)
	assert.Len(t, r.Parameters, 1)
	assert.Equal(t, "p1", r.Parameters[0].Name)
	assert.Equal(t, "参数1", r.Parameters[0].Label)

	// 清理测试数据
	err = o.Delete()
	assert.Nil(t, err)
	exists, _ := GetById(o.ID)
	assert.Nil(t, exists)
}
