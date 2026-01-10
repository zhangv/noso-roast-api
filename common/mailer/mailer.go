package mailer

import (
	"crypto/tls"

	"gopkg.in/gomail.v2"
)

type Mailer struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	UseTLS   bool
}

func NewMailer(host string, port int, username, password, from string, useTLS bool) *Mailer {
	return &Mailer{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
		UseTLS:   useTLS,
	}
}

func (m *Mailer) dialer() *gomail.Dialer {
	d := gomail.NewDialer(m.Host, m.Port, m.Username, m.Password)

	if m.UseTLS {
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return d
}

// 发送消息
func (m *Mailer) Send(msg *Message) error {
	d := m.dialer()
	return d.DialAndSend(msg.m)
}
