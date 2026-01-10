package bean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// 测试创建用户
func TestStocking(t *testing.T) {
	setupTestDB()
	bean := Bean{
		Name:    "test coffee bean",
		Code:    "code",
		Country: "china"}
	_, err := Update(&bean)

	if err != nil {
		t.Fatalf("更新豆子失败: %v", err)
	}

	bean.Stocking(10, BeanTypeGreen, StockLogTypePurchase, "test", "test", "测试入库")

	assert.Equal(t, bean.RawStock, 10)

	bean.Stocking(-10, BeanTypeGreen, StockLogTypeSale, "test", "test", "测试出库")

	assert.Equal(t, bean.RawStock, 0)
}
