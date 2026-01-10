package mailer

import (
	"os"

	"gopkg.in/gomail.v2"
)

type Message struct {
	m *gomail.Message
}

// NewMessage 创建新邮件
func NewMessage() *Message {
	return &Message{
		m: gomail.NewMessage(),
	}
}

// SetFrom 设置发件人
func (m *Message) SetFrom(addr string) {
	m.m.SetHeader("From", addr)
}

// SetTo 设置收件人
func (m *Message) SetTo(addrs ...string) {
	m.m.SetHeader("To", addrs...)
}

// SetSubject 设置主题
func (m *Message) SetSubject(subject string) {
	m.m.SetHeader("Subject", subject)
}

// SetBody 设置纯文本与 HTML
func (m *Message) SetBody(plain string, html string) {
	m.m.SetBody("text/plain", plain)
	m.m.AddAlternative("text/html", html)
}

func (m *Message) SetBodyFromTemplate(templatename string, data interface{}) error {
	body, err := RenderTemplate("templates/"+templatename+".html", data)
	if err != nil {
		return err
	}
	m.SetBody("text/html", body)
	return nil
}

// AddInline 添加 CID 内联图片（从文件路径）
// path: 图片文件路径
// cid: Content-ID，对应模板中的 cid:xxx
// 示例：msg.AddInline("assets/logo.png", "logo")
// 模板中使用：<img src="cid:logo">
func (m *Message) AddInline(path, cid string) {
	m.m.Embed(path, gomail.SetHeader(map[string][]string{
		"Content-ID": {"<" + cid + ">"},
	}))
}

// AddInlineFromBytes 从字节数组添加内联图片
// 注意：gomail 的 Embed 方法只支持文件路径，不支持直接传入字节数组
// 如果需要从内存数据添加图片，请先保存到临时文件，然后使用 AddInline
// data: 图片的字节数据（暂不支持，请使用文件路径方式）
// cid: Content-ID，对应模板中的 cid:xxx
// filename: 临时文件名（用于确定Content-Type），如 "logo.png"
// 示例：先保存图片到文件，然后使用 msg.AddInline("path/to/logo.png", "logo")
func (m *Message) AddInlineFromBytes(data []byte, cid, filename string) error {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "mail_*.png")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name()) // 清理临时文件

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()

	// 使用临时文件路径添加内联图片
	m.AddInline(tmpFile.Name(), cid)
	return nil
}

// AddAttachment 添加附件（正常文件）
func (m *Message) AddAttachment(path string) {
	m.m.Attach(path)
}
