package roast

import (
	"encoding/json"
	"fmt"
	"hahacking/config"
	"hahacking/database"
	"hahacking/models/bean"
	models "hahacking/models/roast"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func Init() { //手动调用
	// config.Setup("../../", "test")
	config.InitViper("test", "../../")
	database.Setup()
	db := database.Conn
	// 自动迁移表结构
	err := db.AutoMigrate(&models.Roast{}, &models.RoastItem{}, &models.RoastProfile{}, &bean.BeanStockLog{},
		&models.RoastRecipe{})
	if err != nil {
		panic("failed to migrate database: " + err.Error())
	}
}

func TestStart(t *testing.T) {
	Init()
	b, _ := bean.Update(&bean.Bean{Name: "Test Bean"})
	bb1, _ := bean.UpdateBatch(&bean.BeanBatch{BeanID: b.ID, TotalQty: 100, BeanType: bean.BeanTypeGreen})
	bb2, _ := bean.UpdateBatch(&bean.BeanBatch{BeanID: b.ID, TotalQty: 100, BeanType: bean.BeanTypeGreen})
	// 创建一个请求
	req, _ := http.NewRequest("POST", "/roast/start", strings.NewReader(`{
	"date":"2025-04-30","inqty":4,
	"name":"test roast","bean_id":`+strconv.Itoa(int(b.ID))+
		`,"items":[{"beanbatch_id":`+strconv.Itoa(int(bb1.ID))+`,"qty":1},
	{"beanbatch_id":`+strconv.Itoa(int(bb2.ID))+`,"qty":3}]}`))
	// 创建一个响应
	w := httptest.NewRecorder()
	// 创建一个上下文
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	Start(c)
	assert.Equal(t, 200, w.Code)
	var responseData map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &responseData)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	data := responseData["data"].(map[string]any)
	roastID := int(data["id"].(float64))
	w.Body.Reset()

	req, _ = http.NewRequest("POST", "/roast/complete", strings.NewReader(
		`{"id":`+strconv.Itoa(int(roastID))+`,"outqty":2,"agtron":2,"agtron_ground":90,"degree":"medium","lossrate":1}`))
	c.Request = req
	Complete(c)
	assert.Equal(t, 200, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &responseData)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	data = responseData["data"].(map[string]any)
	assert.Equal(t, float64(2), data["outqty"])

	//测试保存配方
	roastprofile, _ := models.GetProfileByRoastId(uint(roastID))
	assert.Equal(t, 2, roastprofile.Agtron)
	assert.Equal(t, 90, roastprofile.AgtronGround)
	assert.Equal(t, float64(1), roastprofile.LossRate)
	assert.Equal(t, "medium", string(roastprofile.Degree))

	w.Body.Reset()
	// log.Println("roastIDxxxxxxxxxxxxxxxxxx", roastID, "/roast/saveRecipe/"+strconv.Itoa(int(roastID)))
	req, _ = http.NewRequest("POST", "/roast/"+strconv.Itoa(int(roastID))+"/saverecipe", nil)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: strconv.Itoa(int(roastID))}}
	SaveRecipe(c)
	assert.Equal(t, 200, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &responseData)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	fmt.Println("responsedata", responseData)
	data = responseData["data"].(map[string]any) //recipe object
	// log.Println("1111", b.ID, data["bean_id"], "222", data["roastprofile_id"], data["memo"], data["status"], data["degree"], data["agtron"], data["lossrate"])
	assert.Equal(t, b.ID, uint(data["bean_id"].(float64)))
	// assert.Equal(t, "test", data["memo"])
	assert.Equal(t, "valid", data["status"])
	assert.Equal(t, "medium", string(data["degree"].(string)))
	assert.Equal(t, float64(2), data["agtron"])
	assert.Equal(t, float64(90), data["agtron_ground"])
	assert.Equal(t, float64(1), data["lossrate"])
}
