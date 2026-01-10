package order

import (
	"encoding/json"
	"hahacking/config"
	"hahacking/database"
	"hahacking/models/bean"
	"hahacking/models/order"
	"hahacking/models/shop"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm/logger"
)

func Init() { //手动调用
	config.InitViper("test", "../../")
	database.Setup()
	database.SetLogLevel(logger.Silent)
	// 迁移所有相关表
	database.Conn.AutoMigrate(
		&order.Order{}, &order.OrderItem{}, &order.OrderPayment{}, &order.OrderPaymentItem{},
		&bean.Bean{}, &bean.BeanBatch{}, &bean.BeanBatchLog{}, &bean.BeanStockLog{},
		&shop.ShopAccount{}, &shop.ShopAccountLog{},
	)

	// 创建测试必需的豆子数据
	createTestBeans()
}

func createTestBeans() {
	db := database.Conn

	// 创建测试豆子
	testBeans := []bean.Bean{
		{ID: 1, Name: "Test Bean 1", Status: "valid"},
		{ID: 8, Name: "Test Bean 8", Status: "valid"},
	}

	for _, b := range testBeans {
		var existing bean.Bean
		if err := db.First(&existing, b.ID).Error; err != nil {
			db.Create(&b)
		}
	}

	// 创建测试批次，并确保有正确的 ID
	testBatches := []bean.BeanBatch{
		{BeanID: 1, BeanType: "roasted", Status: "valid", TotalQty: 100, CurrentQty: 100},
		{BeanID: 8, BeanType: "roasted", Status: "valid", TotalQty: 100, CurrentQty: 100},
	}

	for _, batch := range testBatches {
		var existing bean.BeanBatch
		if err := db.Where("bean_id = ? AND bean_type = ?", batch.BeanID, batch.BeanType).First(&existing).Error; err != nil {
			db.Create(&batch)
		}
	}
}

// 获取测试批次的 ID
func getTestBatchID(beanID uint) uint {
	db := database.Conn
	var batch bean.BeanBatch
	if err := db.Where("bean_id = ? AND bean_type = ?", beanID, "roasted").First(&batch).Error; err == nil {
		return batch.ID
	}
	return 0
}
func TestCreate(t *testing.T) {
	Init()
	// 创建一个请求
	req, _ := http.NewRequest("POST", "/order/create", strings.NewReader(`{"totalamt":20,"items":[{"bean_id":1,"qty":1,"amt":12},{"bean_id":8,"qty":3,"amt":8}]}`))
	// 创建一个响应
	w := httptest.NewRecorder()
	// 创建一个上下文
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	Create(c)
	assert.Equal(t, 200, w.Code)
}
func TestPlace(t *testing.T) {
	Init()
	w := httptest.NewRecorder()
	// 创建一个上下文
	c, _ := gin.CreateTestContext(w)
	var req *http.Request

	// 获取测试批次的 ID
	batchID1 := getTestBatchID(1)
	batchID8 := getTestBatchID(8)

	req, _ = http.NewRequest("POST", "/order/create", strings.NewReader(`{"totalamt":21,
	"items":[{"bean_id":1,"beanbatch_id":`+strconv.Itoa(int(batchID1))+`,"qty":1,"amt":12,"price":12},{"bean_id":8,"beanbatch_id":`+strconv.Itoa(int(batchID8))+`,"qty":3,"amt":9,"price":3}]}`))
	c.Request = req
	Update(c)
	assert.Equal(t, 200, w.Code)
	var responseData map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &responseData)
	if err != nil {
		log.Println("response data is invalid", w.Body.String())
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	data := responseData["data"].(map[string]any)
	if data["id"] == nil {
		log.Println("creating order failed", w.Body.String())
		t.Fatalf("Failed to get order id")
	}
	orderID := data["id"].(float64)
	w.Body.Reset()

	req, _ = http.NewRequest("POST", "/order/place", strings.NewReader(`{"order_id":`+strconv.Itoa(int(orderID))+`,"payitems":[{"ids":"1","amt":2,"type":"voucher"},{"ids":"2","amt":3,"type":"account"}]}`))
	c.Request = req
	Place(c)
	assert.Equal(t, 200, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &responseData)

	if err != nil {
		t.Fatalf("Failed to unmarshal place response: %v", err)
	}
	data = responseData["data"].(map[string]any)
	log.Println("placing response", w.Body.String())
	orderPaymentID := data["id"].(float64)
	w.Body.Reset()

	//payment的组合计算是在前端计算完成并提交进来的
	req, _ = http.NewRequest("POST", "/order/settle", strings.NewReader(
		`{"order_id":`+strconv.Itoa(int(orderID))+
			`,"order_payment_id":`+strconv.Itoa(int(orderPaymentID))+
			`,"payment_items":[{"ids":"1","amt":15,"type":"wx"}]}`))
	c.Request = req
	Settle(c)
	assert.Equal(t, 200, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &responseData)
	log.Println("settling aaaaaaaaaaaaaa", w.Body.String())
}
