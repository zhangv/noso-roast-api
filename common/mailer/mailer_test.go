package mailer

import (
	"strings"

	"hahacking/config"
	"testing"
)

// Send 发送邮件
func TestSend(t *testing.T) {
	t.Skip("手动测试")
	config.Setup("../../", config.ENV_TEST)
	cfg := config.Get("mail")

	// 安全地获取配置值，提供默认值
	host, _ := cfg["host"].(string)
	port, _ := cfg["port"].(float64)
	username, _ := cfg["username"].(string)
	password, _ := cfg["password"].(string)
	from, _ := cfg["from"].(string)

	// useTLS 可能不存在，使用默认值 false
	useTLS := false
	if val, ok := cfg["useTLS"].(bool); ok {
		useTLS = val
	}

	mailer := NewMailer(host,
		int(port),
		username,
		password,
		from,
		useTLS,
	)
	msg := NewMessage()
	msg.SetFrom(from)
	msg.SetTo("derekzhangv@gmail.com")
	msg.SetSubject("测试邮件")
	body, err := RenderTemplate("templates/default.html", map[string]interface{}{
		"Content": "测试邮件内容",
		// "URL":     "https://diy.biasura.com",
	})
	if err != nil {
		t.Errorf("渲染模板失败：%v", err)
	}
	msg.SetBody("test", body)

	// 添加内联图片（如果图片文件存在）
	// CID "logo" 对应模板中的 <img src="cid:logo">
	// 图片路径相对于项目根目录或使用绝对路径
	logoPath := "assets/logo.png" // 图片路径
	msg.AddInline(logoPath, "logo")

	err = mailer.Send(msg)
	if err != nil {
		t.Errorf("发送邮件失败：%v", err)
	}
}

// TestRenderTemplateWithoutURL 测试没有URL时的模板渲染
func TestRenderTemplateWithoutURL(t *testing.T) {
	// 测试没有URL的情况
	body, err := RenderTemplate("templates/default.html", map[string]interface{}{
		"Content": "测试邮件内容（无URL）",
		// 不提供 URL 字段
	})
	if err != nil {
		t.Errorf("渲染模板失败：%v", err)
		return
	}

	// 验证按钮不应该出现在HTML中
	if strings.Contains(body, "点击查看") {
		t.Error("当URL不存在时，不应该显示'点击查看'按钮")
	}

	// 测试URL为空字符串的情况
	body2, err := RenderTemplate("templates/default.html", map[string]interface{}{
		"Content": "测试邮件内容（URL为空）",
		"URL":     "",
	})
	if err != nil {
		t.Errorf("渲染模板失败：%v", err)
		return
	}

	if strings.Contains(body2, "点击查看") {
		t.Error("当URL为空字符串时，不应该显示'点击查看'按钮")
	}
}
