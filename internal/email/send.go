package email

import (
	"crypto/tls"
	"fmt"
	"os"
	"strings"

	"gopkg.in/gomail.v2"
)

type Email struct {
	Host                 string
	Port                 int
	Username             string
	Password             string
	From                 string
	verificationTemplate string
}

func (e *Email) Init(paths map[string]string) error {
	fileContent, err := os.ReadFile(paths["verificationTemplate"])
	if err != nil {
		return err
	}
	e.verificationTemplate = string(fileContent)
	return nil
}

func (m *Email) send(msg *gomail.Message) error {
	d := gomail.NewDialer(
		m.Host,
		m.Port,
		m.Username,
		m.Password,
	)
	// 关闭SSL协议认证
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	if err := d.DialAndSend(msg); err != nil {
		return err
	}
	return nil
}

func (e *Email) SendVerification(address, behavior string, verification int) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", e.From)
	msg.SetHeader("To", address)
	msg.SetHeader("Subject", "[OJ++]邮箱验证码")
	template := e.verificationTemplate
	template = strings.Replace(template, "{{behavior}}", behavior, -1)
	template = strings.Replace(template, "{{verification}}", fmt.Sprint(verification), -1)
	msg.SetBody("text/html", template)
	err := e.send(msg)
	return err
}
