package auth

import (
	"hahacking/config"
	"hahacking/database"
	"hahacking/i18n"
	"hahacking/models/user"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm/logger"
)

func Init() { //手动调用
	// config.InitViper("test", "../../")
	config.Setup("../../", "test")
	database.Setup()
	database.SetLogLevel(logger.Info)
	database.Conn.AutoMigrate(&user.User{})
	i18n.Init() //初始化i18n - 会依赖
}
func TestRegister(t *testing.T) {
	Init()
	// 创建一个请求
	req, _ := http.NewRequest("POST", "/auth/register", strings.NewReader(`{"nickname":"test","email":"test@test.com","password":"123456"}`))
	// 创建一个响应
	w := httptest.NewRecorder()
	// 创建一个上下文
	c, _ := gin.CreateTestContext(w)
	i18n.SetupContext(c, "zh-Hans") //根据需要设置
	c.Request = req

	Register(c)
	assert.Equal(t, 200, w.Code)
}
