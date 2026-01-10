package user

import (
	"encoding/json"
	"hahacking/config"
	"hahacking/database"
	models "hahacking/models/user"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm/logger"
)

func Init() { //手动调用
	config.InitViper("test", "../../")
	database.Setup()
	database.SetLogLevel(logger.Info)
	database.Conn.AutoMigrate(&models.User{})
}
func TestUpdate(t *testing.T) {
	Init()
	var responseData map[string]any
	email := time.Now().Format("20060102150405") + "@test.com"
	// 创建一个请求
	req, _ := http.NewRequest("POST", "/user/update", strings.NewReader(`{"nickname":"test","email":"`+email+`","type":"user","status":"active"}`))
	// 创建一个响应
	w := httptest.NewRecorder()
	// 创建一个上下文
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	Update(c)
	assert.Equal(t, 200, w.Code)
	// body = w.Body.String()
	// log.Println(body) //TODO 可能报错duplicate mobile column
	err := json.Unmarshal(w.Body.Bytes(), &responseData)
	if err != nil {
		log.Println("response data is invalid", w.Body.String())
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	// data := responseData["data"].(map[string]any)
	w.Body.Reset()

	req, _ = http.NewRequest("GET", "/user/search?nickname=test&email="+email, nil)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	Search(c)
	assert.Equal(t, 200, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &responseData)
	if err != nil {
		log.Println("response data is invalid", w.Body.String())
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	// log.Println("aaaaaaaaaaaaa", responseData)
	// body := w.Body.String()
	w.Body.Reset()
	// log.Println("body", body)
	data := responseData["data"].(map[string]any)
	// 获取 objects 字段
	objects := data["objects"].([]any)
	// log.Println("xxx", objects)
	first := objects[0].(map[string]any)
	assert.Equal(t, "test", first["nickname"])
	assert.Equal(t, email, first["email"].(string))
	assert.Equal(t, 1, len(objects))
	// if ok {
	// 	// objects 是一个 interface{}，通常是 []interface{}
	// 	// 你可以断言为切片
	// 	objSlice, ok := objects.([]any)
	// 	if ok && len(objSlice) > 0 {
	// 		firstObj := objSlice[0]
	// 		log.Println("ttttttttttt", firstObj)
	// 		// assert.Equal(t, "test", firstObj["nickname"])
	// 		// assert.Equal(t, "test@test.com", firstObj["email"])
	// 		// assert.Equal(t, "13345678901", firstObj["mobile"])
	// 		// assert.Equal(t, "user", firstObj["type"])
	// 		// assert.Equal(t, "active", firstObj["status"])
	// 	}
	// } else {
	// 	assert.Fail(t, "objects 字段不存在")
	// }
}
