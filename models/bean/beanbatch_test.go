package bean

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 测试创建用户
func TestUpdateBatch(t *testing.T) {
	setupTestDB()
	bean := Bean{
		Name:         "test coffee bean" + time.Now().Format("2006-01-02 15:04"),
		Code:         "code",
		Country:      "china",
		RawStock:     0,
		RoastedStock: 0,
	}
	Update(&bean)

	batch := BeanBatch{
		BeanID:     bean.ID,
		Seq:        1,
		Date:       "2021-01-01",
		BeanType:   BeanTypeGreen,
		Status:     BatchStatusValid,
		TotalPrice: 100,
		TotalQty:   100,
	}
	UpdateBatch(&batch)
	var newbean *Bean
	newbean, _ = GetById(bean.ID)
	assert.Equal(t, 100, newbean.RawStock)

	err := batch.Undo()
	if err != nil {
		t.Fatalf("撤销批次失败: %v", err)
	}

	var updatedBean *Bean
	updatedBean, err = GetById(bean.ID)
	if err != nil {
		t.Fatalf("获取豆子失败: %v", err)
	}
	assert.Equal(t, 0, updatedBean.RawStock)

}

func TestLog(t *testing.T) {
	setupTestDB()
	_, batch, err := NewBean()
	assert.NoError(t, err)
	batch.Log(10, "user", "1", "")
	assert.Equal(t, 110, batch.CurrentQty)
	logs, _ := SearchBatchLog(SearchBatchLogForm{
		BeanBatchID: batch.ID,
	})
	assert.Equal(t, 1, len(logs))
	assert.Equal(t, 10, logs[0].Qty)
	assert.Equal(t, "user", logs[0].RefType)
	assert.Equal(t, "1", logs[0].RefID)

	batch.Log(10, "user", "1", "")
	assert.Equal(t, 120, batch.CurrentQty)

	batch.Log(10, "NOTUSER", "1", "")
	assert.Equal(t, 130, batch.CurrentQty)

	err = batch.Log(10, "NOTUSER", "1", "")
	assert.Error(t, err) //重复的ref - 报错
	assert.Equal(t, 130, batch.CurrentQty)
}
