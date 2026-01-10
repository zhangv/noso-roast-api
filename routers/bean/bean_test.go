package bean

import (
	"encoding/json"
	"fmt"
	"hahacking/config"
	"hahacking/database"
	models "hahacking/models/bean"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func Init() { //手动调用
	// config.Setup("../../", "test")
	config.InitViper("test", "../../")
	database.Setup()
	db := database.Conn
	// 自动迁移表结构
	err := db.AutoMigrate(&models.Bean{}, &models.BeanBatch{}, &models.BeanStockLog{},
		&models.BeanBatchLog{})
	if err != nil {
		panic("failed to migrate database: " + err.Error())
	}
}

func TestUpdate(t *testing.T) {
	Init()
	//更新已存在的对象
	b, _ := models.Update(&models.Bean{Name: "Test Bean", Code: "testcode" + time.Now().Format("20060102150405")})
	// 创建一个请求
	req, _ := http.NewRequest("POST", "/bean/update", strings.NewReader(`{
	"id":`+strconv.Itoa(int(b.ID))+
		`,"code":"`+b.Code+`",`+
		`"name":"test bean updated","status":"valid",
		"type":"green","variety":"test variety","country":"test country",
		"processor":"test processor","process":"test process","provider":"test provider",
		"purchaseprice":100,"cprice":120,"bprice":100,"owner":"test owner","description":"test description",
		"rawstockalert":100,"roastedstockalert":100}`))
	// 创建一个响应
	w := httptest.NewRecorder()
	// 创建一个上下文
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	Update(c)
	assert.Equal(t, 200, w.Code)
	var responseData map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &responseData)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	data := responseData["data"].(map[string]any)
	beanID := int(data["id"].(float64))
	w.Body.Reset()

	assert.Equal(t, b.ID, uint(beanID))
	assert.Equal(t, b.Code, data["code"])
	assert.Equal(t, "valid", data["status"])
	assert.Equal(t, "green", data["type"])
	assert.Equal(t, "test variety", data["variety"])
	assert.Equal(t, "test country", data["country"])
	assert.Equal(t, "test processor", data["processor"])
}
func TestSearch(t *testing.T) {
	Init()
	b, _ := models.Update(&models.Bean{Name: "Test Bean search", Code: "testcodesearch" + time.Now().Format("20060102150405")})
	//测试搜索
	req, _ := http.NewRequest("GET", "/bean/search?code="+b.Code, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	Search(c)
	assert.Equal(t, 200, c.Writer.Status())
	fmt.Println(w.Body.String())
	responseData := map[string]any{}
	err := json.Unmarshal(w.Body.Bytes(), &responseData)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	data := responseData["data"].(map[string]any)
	objects := data["objects"].([]any)
	total := data["total"].(float64)
	assert.Equal(t, 1, int(total))
	assert.Equal(t, 1, len(objects))
	bean := objects[0].(map[string]any)
	assert.Equal(t, b.Name, bean["name"])
	assert.Equal(t, b.Code, bean["code"])
	assert.Equal(t, b.Status, bean["status"])
	assert.Equal(t, string(b.Type), bean["type"])
}
