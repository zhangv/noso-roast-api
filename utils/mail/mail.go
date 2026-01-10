package mail

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"hahacking/config"
	"io"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"net/url"
	"strconv"

	"github.com/resend/resend-go/v2"
	"gopkg.in/gomail.v2"
)

func Send(to string, subject string, body string) bool {
	if config.ENV == config.ENV_PRD {
		// return Send_sendcloud(to, subject, body)
		from := "v@resend.hahacking.com"
		fromName := "ZhangV"
		return Send_resend(to, subject, body, from, fromName)
	} else {
		log.Println("sendmail in dev or test env")
		return true
	}
}
func Send_sendcloud(to string, subject string, body string) bool {
	conf := config.GetConfig()
	sendcloud := conf.Get("sendcloud").(map[string]any)
	api_user := sendcloud["api_user"].(string)
	api_key := sendcloud["api_key"].(string)

	RequestURI := "http://api.sendcloud.net/apiv2/mail/send"
	PostParams := url.Values{
		"apiUser":  {api_user},
		"apiKey":   {api_key},
		"from":     {"cs@mailer.weordos.com"},
		"fromName": {"goggo"},
		"to":       {to},
		"subject":  {subject},
		"html":     {body},
	}
	PostBody := bytes.NewBufferString(PostParams.Encode())

	ResponseHandler, err := http.Post(RequestURI, "application/x-www-form-urlencoded", PostBody)
	if err != nil {
		log.Println(err)
		return false
	}
	defer ResponseHandler.Body.Close()
	BodyByte, err := io.ReadAll(ResponseHandler.Body)
	if err != nil {
		log.Println(err)
		return false
	}
	log.Println(string(BodyByte))
	return true
}

func Send_resend(to string, subject string, body string, from string, fromName string) bool {
	apiKey := "re_eoZGPvN4_KtbsmtXJDcfAM1CAQZnBa3n7" //配额：每天100，每月3000

	client := resend.NewClient(apiKey)

	fromEmail := fromName + "<" + from + ">"
	params := &resend.SendEmailRequest{
		From:    fromEmail,
		To:      []string{to},
		Subject: subject,
		Html:    body,

		// From:    "Acme <onboarding@resend.dev>",
		// To:      []string{"delivered@resend.dev"},
		// Html:    "<strong>hello world</strong>",
		// Subject: "Hello from Golang",
		// Cc:      []string{"cc@example.com"},
		// Bcc:     []string{"bcc@example.com"},
		// ReplyTo: "replyto@example.com",
	}

	sent, err := client.Emails.Send(params)
	if err != nil {
		log.Println(err)
		return false
	}
	log.Println(sent)
	return true
}

func Send_smtp(to string, subject string, body string) {
	host := "smtp.exmail.qq.com"
	port := 25
	password := "mNnusGnjeAR6tZQU"

	header := make(map[string]string)
	email := "cs@weordos.com"
	header["From"] = "Hacking " + "<" + email + ">"
	header["To"] = to
	header["Subject"] = "Dockboard 帐户激活邮件"
	header["Content-Type"] = "text/html; charset=UTF-8"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	auth := smtp.PlainAuth(
		"",
		email,
		password,
		host,
	)

	err := SendMailUsingTLS(
		fmt.Sprintf("%s:%d", host, port),
		auth,
		email,
		[]string{to},
		[]byte(message),
	)

	if err != nil {
		panic(err)
	}
}

// return a smtp client
func Dial(addr string) (*smtp.Client, error) {
	conn, err := tls.Dial("tcp", addr, nil)
	if err != nil {
		log.Println("Dialing Error:", err)
		return nil, err
	}
	//分解主机端口字符串
	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}

// 参考net/smtp的func SendMail()
// 使用net.Dial连接tls(ssl)端口时,smtp.NewClient()会卡住且不提示err
// len(to)>1时,to[1]开始提示是密送
func SendMailUsingTLS(addr string, auth smtp.Auth, from string,
	to []string, msg []byte) (err error) {

	//create smtp client
	c, err := Dial(addr)
	if err != nil {
		log.Println("Create smpt client error:", err)
		return err
	}
	defer c.Close()

	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(auth); err != nil {
				log.Println("Error during AUTH", err)
				return err
			}
		}
	}

	if err = c.Mail(from); err != nil {
		return err
	}

	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()
}

func Send_qq_exmail(to string, subject string, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "cs@weordos.com")
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	// m.Attach("/home/Alex/lolcat.jpg")

	d := gomail.NewDialer("smtp.exmail.qq.com", 587, "cs@weordos.com", "mNnusGnjeAR6tZQU")
	d.SSL = true
	d.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	// d.Auth = smtp.PlainAuth("", "cs@weordos.com", "mNnusGnjeAR6tZQU", "smtp.exmail.qq.com")

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}

	return nil
}
func Send0(to string, subject string, body string) error {
	conf := config.GetConfig()
	log.Println("conf", conf)
	smtpCfg := conf.Get("smtp").(map[string]interface{})
	cs := smtpCfg["cs"].(map[string]interface{})
	// notify := smtpCfg["notify"].(map[string]interface{})
	log.Println("xxxxxxxxxcs", cs)
	smtpHost := cs["host"].(string)
	smtpPort := cs["port"].(float64)
	senderEmail := cs["username"].(string)
	senderPassword := cs["password"].(string) // 使用QQ邮箱的授权码
	recipientEmail := to

	// 邮件内容
	message := []byte("To: " + recipientEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	// 认证信息
	auth := smtp.PlainAuth("", senderEmail, senderPassword, smtpHost)
	// 发送邮件
	err := smtp.SendMail(smtpHost+":"+strconv.Itoa(int(smtpPort)), auth, senderEmail, []string{recipientEmail}, message)
	if err != nil {
		log.Fatalf("Failed to send email: %v", err)
	}

	log.Println("Email sent successfully!")
	return nil
}
