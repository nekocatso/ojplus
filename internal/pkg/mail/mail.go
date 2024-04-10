package mail

import (
	"errors"

	"gopkg.in/gomail.v2"
)

type MailBox struct {
	name     string
	password string
	host     string
	port     int
}

type MailPool struct {
	sum int
	num int
	mailbox []MailBox
}

func NewMailPool(name []string, password []string, host []string, port []int) (*MailPool,error) {
	var mp MailPool
	if(len(name) != len(password) || len(name) != len(host) || len(name) != len(port)){
		return nil,errors.New("参数长度不匹配")
	}
	for i := 0; i < len(name); i++ {
		mp.mailbox = append(mp.mailbox, MailBox{name: name[i], password: password[i], host: host[i], port: port[i]})
	}
	mp.sum = len(name)
	mp.num = 0
	return &mp,nil
}
func (mp *MailPool) SendMail(subject string, to []string, Cc []string, Bcc []string, message string, annex []string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", mp.mailbox[mp.num].name)
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
		mp.mailbox[mp.num].host, mp.mailbox[mp.num].port, mp.mailbox[mp.num].name, mp.mailbox[mp.num].password,
	)
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
