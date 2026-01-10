package bean

import (
	"hahacking/config"
	"hahacking/database"
	"hahacking/models/price"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func setupTestDB() *gorm.DB {
	config.InitViper("test", "../../")
	database.SetLogLevel(gormLogger.Info)
	database.Setup()
	db := database.GetConn()

	// 自动迁移表结构
	db.AutoMigrate(&Bean{}, &BeanBatch{}, &BeanStockLog{})
	return db
}

// 测试创建用户
func TestUpdate(t *testing.T) {
	setupTestDB()
	bean := Bean{
		Name:    "test coffee bean",
		Code:    "code",
		Country: "china"}
	_, err := Update(&bean)

	if err != nil {
		t.Fatalf("更新豆子失败: %v", err)
	}

	// 检查 ID 是否被 Gorm 赋值
	if bean.ID == 0 {
		t.Fatalf("用户 ID 应该被 Gorm 赋值")
	}
	assert.NotNil(t, bean.ID)
}

func TestSearch(t *testing.T) {
	setupTestDB()
	beans, _ := Search(SearchForm{
		Name: "test",
	})
	assert.NotNil(t, beans)
}

func TestGeneratePrice(t *testing.T) {
	setupTestDB()
	bean, _, _ := NewBean() //已经生成价目表
	bean.CPrice = 100
	Update(bean)
	levels, _ := price.SearchLevel(price.SearchLevelForm{Status: "valid"})

	bean.GeneratePrices(true)
	pl, _ := bean.GetPriceList("")
	assert.Equal(t, len(levels), len(pl))

}

func TestGetPriceList(t *testing.T) {
	setupTestDB()
	beans, _ := GetAllPriceList(price.LevelTypeC)
	assert.NotNil(t, beans)
}
