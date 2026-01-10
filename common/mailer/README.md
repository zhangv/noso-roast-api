# 邮件服务使用指南

## 在邮件模板中添加图片

### 方法一：使用 CID 内联图片（推荐）

这是最常用的方式，图片会嵌入到邮件中，不需要外部链接。

#### 1. 准备图片文件

将图片文件放在 `service/mail/assets/` 目录下，例如：

```
service/mail/assets/logo.png
```

#### 2. 在模板中引用图片

在 HTML 模板中使用 `cid:图片ID` 引用图片：

```html
<!-- templates/default.html -->
<img src="cid:logo" width="80" alt="Logo" />
```

**注意**：CID 值（如 `logo`）必须与代码中的 CID 参数一致。

#### 3. 在发送邮件前添加图片

在发送邮件前，使用 `AddInline` 方法添加图片：

```go
msg := NewMessage()
msg.SetFrom("sender@example.com")
msg.SetTo("recipient@example.com")
msg.SetSubject("测试邮件")

// 渲染模板
body, _ := RenderTemplate("templates/default.html", data)
msg.SetBody("纯文本内容", body)

// 添加内联图片
// 参数1: 图片文件路径（相对于项目根目录或绝对路径）
// 参数2: CID值，必须与模板中的 cid:xxx 一致
msg.AddInline("service/mail/assets/logo.png", "logo")

// 发送邮件
mailer.Send(msg)
```

#### 完整示例

```go
package main

import (
    "hahacking/service/mail"
    "hahacking/config"
)

func sendEmailWithImage() {
    // 初始化配置
    config.Setup(".", config.ENV_DEV)
    cfg := config.Get("mail")

    // 创建邮件发送器
    mailer := mail.NewMailer(
        cfg["host"].(string),
        int(cfg["port"].(float64)),
        cfg["username"].(string),
        cfg["password"].(string),
        cfg["from"].(string),
        cfg["useTLS"].(bool),
    )

    // 创建邮件消息
    msg := mail.NewMessage()
    msg.SetFrom(cfg["from"].(string))
    msg.SetTo("recipient@example.com")
    msg.SetSubject("带图片的邮件")

    // 渲染模板
    body, _ := mail.RenderTemplate("service/mail/templates/default.html", map[string]interface{}{
        "Content": "这是一封带图片的测试邮件",
        "URL":     "https://example.com",
    })
    msg.SetBody("纯文本内容", body)

    // 添加内联图片
    msg.AddInline("service/mail/assets/logo.png", "logo")

    // 发送邮件
    if err := mailer.Send(msg); err != nil {
        log.Printf("发送邮件失败: %v", err)
    }
}
```

### 方法二：使用外部 URL（不推荐）

某些邮件客户端可能会阻止显示外部图片。

```html
<img src="https://example.com/images/logo.png" width="80" alt="Logo" />
```

### 方法三：从字节数组添加图片

如果需要从内存中的字节数据添加图片：

```go
imageBytes := []byte{...} // 图片的字节数据
err := msg.AddInlineFromBytes(imageBytes, "logo", "logo.png")
if err != nil {
    log.Printf("添加图片失败: %v", err)
}
```

**注意**：此方法会创建临时文件，使用后会自动清理。

## 添加多个图片

可以添加多个内联图片，每个图片使用不同的 CID：

```go
msg.AddInline("service/mail/assets/logo.png", "logo")
msg.AddInline("service/mail/assets/banner.jpg", "banner")
msg.AddInline("service/mail/assets/footer.png", "footer")
```

模板中对应使用：

```html
<img src="cid:logo" alt="Logo" />
<img src="cid:banner" alt="Banner" />
<img src="cid:footer" alt="Footer" />
```

## 添加附件

如果需要添加普通附件（非内联图片）：

```go
msg.AddAttachment("path/to/document.pdf")
```

## 注意事项

1. **CID 值必须一致**：模板中的 `cid:xxx` 必须与 `AddInline` 的第二个参数完全一致
2. **图片路径**：可以使用相对路径（相对于项目根目录）或绝对路径
3. **图片格式**：支持 PNG、JPG、GIF 等常见格式
4. **图片大小**：建议图片大小不超过 200KB，以确保邮件发送速度
5. **图片不存在**：如果图片文件不存在，`AddInline` 不会报错，但邮件中图片不会显示

## 常见问题

### Q: 图片不显示怎么办？

A: 检查以下几点：

- CID 值是否一致（模板中的 `cid:logo` 和代码中的 `"logo"`）
- 图片路径是否正确
- 图片文件是否存在
- 邮件客户端是否支持内联图片（大多数现代邮件客户端都支持）

### Q: 可以使用相对路径吗？

A: 可以，路径相对于项目根目录。例如：

- `service/mail/assets/logo.png` ✅
- `./service/mail/assets/logo.png` ✅
- `/absolute/path/to/logo.png` ✅

### Q: 如何添加多个图片？

A: 多次调用 `AddInline`，使用不同的 CID 值即可。
