package i18n

import (
	"encoding/json"
	"fmt"
	"hahacking/config"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// 全局i18n Bundle
var bundle *i18n.Bundle
var C *gin.Context

// 初始化i18n
func Init() {
	// 创建Bundle，设置默认语言
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// 加载翻译文件
	err := filepath.Walk(config.C.BasePath+"/config/locales", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		_, err = bundle.LoadMessageFile(path)
		return err
	})
	if err != nil {
		panic(fmt.Sprintf("加载翻译文件失败: %v", err))
	}
}

// i18n中间件
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Accept-Language头获取语言
		acceptLang := c.GetHeader("Accept-Language")
		lang, _ := language.Parse(acceptLang)
		if lang == language.Und { // 未定义语言时使用默认值
			lang = language.English
		}
		// log.Println("lang", lang)
		// 创建Localizer
		SetupContext(c, lang.String())
		c.Next()
	}
}

// 增加这个方法便于单元测试
func SetupContext(c *gin.Context, lang string) {
	localizer := i18n.NewLocalizer(bundle, lang)
	// log.Println("localizerxxxxxxxxxxx", localizer)
	C = c
	c.Set("i18n", localizer)
}

func T(messageID string, args ...string) string {
	localizer := C.MustGet("i18n").(*i18n.Localizer)
	msg, _ := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: messageID,
		// TemplateData: data,
	})
	if len(args) > 0 {
		for k, v := range args {
			msg = strings.ReplaceAll(msg, "{"+strconv.Itoa(k)+"}", v)
		}
	}
	return msg
}

func T0(c *gin.Context) func(string, ...interface{}) string {
	localizer := c.MustGet("i18n").(*i18n.Localizer)
	return func(messageID string, data ...interface{}) string {
		var templateData map[string]interface{}
		if len(data) > 0 {
			if m, ok := data[0].(map[string]interface{}); ok {
				templateData = m
			}
		}
		msg, err := localizer.Localize(&i18n.LocalizeConfig{
			MessageID:    messageID,
			TemplateData: templateData,
		})
		if err != nil {
			return messageID // 翻译失败时返回ID
		}
		return msg
	}
}
