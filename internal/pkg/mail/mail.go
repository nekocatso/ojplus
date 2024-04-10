package mail

import (
	"gopkg.in/gomail.v2"
)

type MailBox struct {
	name     string
	password string
	host     string
	port     int
}

type MaillPool struct {
	mailmap map[string]string
	mailbox *MailBox
}

func NewMailBox(name string, password string, host string, port int) *MailBox {
	return &MailBox{name: name, password: password, host: host, port: port}
}
func (from *MailBox) SendMail(subject string, to []string, Cc []string, Bcc []string, message string, annex []string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", from.name)
	m.SetHeader("Subject", subject)
	for _, t := range to {

		m.SetHeader("To", t)
	}
	for _, c := range Cc {

		m.SetHeader("Cc", c)
	}
	for _, b := range Bcc {

		m.SetHeader("Bcc", b)
	}
	m.SetBody("text/html", message)
	for _, a := range annex {

		m.Attach(a)
	}
	d := gomail.NewDialer(
		from.host, from.port, from.name, from.password,
	)
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

func NewMailPool(name string, password string, host string, port int) MaillPool {

	return MaillPool{mailbox: NewMailBox(name, password, host, port), mailmap: map[string]string{}}
}
