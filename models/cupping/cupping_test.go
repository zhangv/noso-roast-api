package cupping

import (
	"hahacking/config"
	"hahacking/database"
	"hahacking/models/bean"
	"hahacking/models/roast"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func setupTestDB() *gorm.DB {
	if database.Conn == nil {
		config.Setup("../../", "test")
		database.SetLogLevel(gormLogger.Silent)
		database.Setup()
	}
	db := database.Conn
	// 自动迁移表结构
	db.AutoMigrate(&Cupping{}, &CuppingSample{}, &CuppingScore{}, &bean.BeanStockLog{}, &roast.Roast{}, &roast.RoastItem{}, &roast.RoastProfile{})
	return db
}

func TestNewCupping(t *testing.T) {
	setupTestDB()
	b, batch, _ := bean.NewBean()
	ro, roastedBatch, _ := roast.NewRoast(b, batch)
	cp, _ := NewCupping(b, roastedBatch, ro)
	assert.NotNil(t, cp.ID)
}
