package audit

import (
	"hahacking/config"
	"hahacking/database"
	"hahacking/models/bean"
	"hahacking/models/order"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func setupTestDB() *gorm.DB {
	config.InitViper("test", "../../")
	database.SetLogLevel(gormLogger.Silent)
	database.Setup()
	db := database.GetConn()
	db.AutoMigrate(&AuditLog{})
	// 清理之前的测试数据，确保测试独立性
	db.Exec("DELETE FROM t_audit_log")
	RegisterAuditHooks(db)
	return db
}

func TestUpdateByMap(t *testing.T) {
	setupTestDB()
	b, _ := bean.Update(&bean.Bean{
		Name:    "test coffee bean",
		Code:    "code",
		Country: "china",
	})
	batch, _ := bean.UpdateBatch(&bean.BeanBatch{
		BeanID:     b.ID,
		TotalQty:   10,
		TotalPrice: 100,
		BeanType:   bean.BeanTypeRoasted,
		Type:       bean.StockLogType(bean.BatchTypeProduce),
	})

	orderForm := order.UpdateForm{
		Date:     time.Now().Format("2006-01-02"),
		TotalAmt: 100,
		TotalQty: 1,
		Items: []order.OrderItem{
			{BeanID: b.ID,
				Qty:         1,
				Price:       100,
				Amt:         100,
				BeanBatchID: batch.ID,
				BeanName:    b.Name,
			},
		},
	}
	o, err := order.Update(&orderForm)

	assert.Nil(t, err)

	_, count := Search(SearchForm{
		ObjectName: "Order",
		RecordID:   o.ID,
		Action:     "CREATE",
	})
	assert.Equal(t, 0, count)
	//TODO 删除全部日志数据
}

func TestUpdate(t *testing.T) {
	setupTestDB()
	b := bean.Bean{
		Name:    "test coffee bean",
		Code:    "code",
		Country: "china"}
	_, err := bean.Update(&b)

	assert.Nil(t, err)

	logs, count := Search(SearchForm{
		ObjectName: "Bean",
		RecordID:   b.ID,
		Action:     "CREATE",
	})
	assert.Equal(t, 1, count)
	assert.Equal(t, "CREATE", logs[0].Action)

	b.Status = "valid"
	bean.Update(&b)

	logs, count = Search(SearchForm{
		ObjectName: "Bean",
		RecordID:   b.ID,
		Action:     "UPDATE",
	})
	log.Printf("logs update: %+v", logs)
	assert.Equal(t, 1, count)
	assert.Equal(t, "UPDATE", logs[0].Action)

}
