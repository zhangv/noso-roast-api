package roast

import (
	"hahacking/config"
	"hahacking/database"
	"hahacking/models/bean"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupTestDB() *gorm.DB {
	config.InitViper("test", "../../")
	database.SetLogLevel(gormlogger.Silent)
	database.Setup()
	db := database.Conn
	// 自动迁移表结构
	err := db.AutoMigrate(&Roast{}, &RoastItem{}, &RoastProfile{}, &bean.BeanStockLog{})
	if err != nil {
		panic("failed to migrate database: " + err.Error())
	}
	return db
}

func TestComplete(t *testing.T) {
	setupTestDB()
	b := bean.Bean{
		Name:    "test coffee bean" + time.Now().Format("2006-01-02 15:04"),
		Code:    "code",
		Country: "china"}
	bean.Update(&b)

	batch, err := bean.UpdateBatch(&bean.BeanBatch{
		BeanID:     b.ID,
		Type:       bean.StockLogTypePurchase,
		BeanType:   bean.BeanTypeGreen,
		TotalQty:   100,
		TotalPrice: 100,
	})
	if err != nil {
		t.Fatalf("更新批次失败: %v", err)
	}
	assert.Equal(t, 100, batch.TotalQty)
	assert.Equal(t, 100, batch.CurrentQty)

	items := []RoastItem{
		{
			BeanBatchID: batch.ID,
			Qty:         50,
		},
	}
	startForm := StartForm{
		Name:   "test coffee roast",
		BeanID: b.ID,
		InQty:  50,
		Items:  items,
	}
	roast, err := Start(startForm)
	if err != nil {
		t.Fatalf("烘焙开始失败: %v", err)
	}

	roast, err = GetById(roast.ID)
	if err != nil {
		t.Fatalf("获取烘焙记录失败: %v", err)
	}

	completeForm := CompleteForm{
		ID:           roast.ID,
		OutQty:       40,
		Agtron:       100,
		AgtronGround: 90,
		LossRate:     11.9,
		Degree:       string(RoastDegreeMedium),
	}
	err = roast.Complete(completeForm)
	if err != nil {
		t.Fatalf("完成烘焙失败: %v", err)
	}
	//检查profile是否建立
	profile, _ := GetProfileByRoastId(roast.ID)
	assert.Equal(t, 100, profile.Agtron)
	assert.Equal(t, 90, profile.AgtronGround)
	assert.Equal(t, RoastDegreeMedium, profile.Degree)
	assert.Equal(t, 11.9, profile.LossRate)

	//检测熟豆批次是否更新
	assert.Equal(t, 40, roast.OutQty)
	roastedBatch, err := bean.GetBatchById(roast.BeanBatchID)
	if err != nil {
		t.Fatalf("获取熟豆批次失败: %v", err)
	}
	assert.Equal(t, 40, roastedBatch.CurrentQty)

	//检测生豆批次是否更新
	greenBatch, err := bean.GetBatchById(batch.ID)
	if err != nil {
		t.Fatalf("获取生豆批次失败: %v", err)
	}
	assert.Equal(t, 50, greenBatch.CurrentQty)
}
