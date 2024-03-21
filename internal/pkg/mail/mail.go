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
type Mail struct {
	mailbox *MailBox
	message *gomail.Message
}

func NewMailBox(name string, password string, host string, port int) *MailBox {
	return &MailBox{name: name, password: password, host: host, port: port}
}
func SendMail(from *MailBox, subject string, to []string, Cc []string, Bcc []string, message string, annex []string) error {
	m := &Mail{mailbox: from, message: gomail.NewMessage()}
	m.message.SetHeader("From", from.name)
	m.message.SetHeader("Subject", subject)
	for _, t := range to {

		m.message.SetHeader("To", t)
	}
	for _, c := range Cc {

		m.message.SetHeader("Cc", c)
	}
	for _, b := range Bcc {

		m.message.SetHeader("Bcc", b)
	}
	m.message.SetBody("text/html", message)
	for _, a := range annex {

		m.message.Attach(a)
	}
	d := gomail.NewDialer(
		from.host, from.port, from.name, from.password,
	)
	if err := d.DialAndSend(m.message); err != nil {
		return err
	}

	return nil

}
