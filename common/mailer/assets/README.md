# 邮件图片资源目录

此目录用于存放邮件模板中使用的图片文件。

## 使用方法

1. 将图片文件（如 logo.png）放在此目录下
2. 在邮件模板中使用 `cid:图片ID` 引用：
    ```html
    <img src="cid:logo" width="80" />
    ```
3. 在发送邮件前，使用 `AddInline` 方法添加图片：
    ```go
    msg.AddInline("service/mail/assets/logo.png", "logo")
    ```

## 注意事项

- CID（Content-ID）必须与模板中的 `cid:` 后面的值一致
- 图片路径可以是相对路径（相对于项目根目录）或绝对路径
- 支持的图片格式：PNG、JPG、GIF 等常见格式
- 建议图片大小不超过 200KB，以确保邮件发送速度
