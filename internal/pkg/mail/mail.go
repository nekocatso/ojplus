package mail

import (
	"errors"

	"gopkg.in/gomail.v2"
)

// MailBox
// 功能：定义邮箱账户信息结构体，包含邮箱地址、密码、smtp服务器地址、端口号等属性
//
// 字段：
//
//	name string // 邮箱地址，字符串类型
//	password string // 密码，字符串类型
//	host string // smtp服务器地址，字符串类型，指定邮件服务器的域名或IP
//	port int // 端口号，整数类型，指定邮件服务器的服务端口
type MailBox struct {
	name     string // 邮箱地址
	password string // 密码
	host     string // smtp服务器地址
	port     int    // 端口号
}

// MailPool
// 功能：定义邮箱池结构体，包含邮箱总数、当前使用邮箱以及一个包含多个MailBox的切片
//
// 字段：
//
//	sum int // 总邮箱数，整数类型
//	num int // 当前使用邮箱，整数类型
//	mailbox []MailBox // 邮箱账户切片，存储一组MailBox实例
type MailPool struct {
	sum     int       // 总邮箱数
	num     int       // 当前使用邮箱数
	mailbox []MailBox // 邮箱账户切片
}

// NewMailPool
// 功能：创建并初始化一个新的MailPool实例，根据提供的邮箱账户信息数组构建邮箱池
//
// 参数：
//
//	name []string // 邮箱地址数组，每个元素对应一个邮箱账户的邮箱地址
//	password []string // 密码数组，每个元素对应一个邮箱账户的密码
//	host []string // smtp服务器地址数组，每个元素对应一个邮箱账户的smtp服务器地址
//	port []int // 端口号数组，每个元素对应一个邮箱账户的端口号
//
// 返回值：
//
//	*MailPool // 初始化后的MailPool指针，若成功则返回新创建的邮箱池实例，否则返回nil
//	error // 错误信息，若创建过程中出现错误则返回相应的错误信息，否则返回nil
func NewMailPool(name []string, password []string, host []string, port []int) (*MailPool, error) {
	// 创建空MailPool实例
	var mp MailPool

	// 检查输入参数长度是否一致，如果不一致则返回错误
	if len(name) != len(password) || len(name) != len(host) || len(name) != len(port) {
		return nil, errors.New("参数长度不匹配")
	}

	// 根据输入参数创建并添加MailBox到mp.mailbox切片中
	for i := 0; i < len(name); i++ {
		mp.mailbox = append(mp.mailbox, MailBox{
			name:     name[i],     // 邮箱地址
			password: password[i], // 密码
			host:     host[i],     // smtp服务器地址
			port:     port[i],     // 端口号
		})
	}

	// 设置mp.sum为输入邮箱地址数组长度，表示总邮箱数
	mp.sum = len(name)

	// 初始化mp.num为0，表示当前可用邮箱数
	mp.num = 0

	// 返回新创建的MailPool实例及其指针
	return &mp, nil
}

// SendMail
// 功能：使用MailPool中的邮箱账户发送一封邮件
//
// 参数：
//   subject string // 邮件主题，字符串类型
//   to []string // 收件人列表，字符串数组，每个元素为一个收件人的电子邮件地址
//   Cc []string // 抄送人列表，字符串数组，每个元素为一个抄送人的电子邮件地址
//   Bcc []string // 密送人列表，字符串数组，每个元素为一个密送人的电子邮件地址
//   message string // 邮件正文，字符串类型，包含HTML格式的邮件内容
//   annex []string // 附件路径列表，字符串数组，每个元素为一个本地文件路径，这些文件将作为附件添加到邮件中
//
// 返回值：
//   error // 错误信息，若发送邮件过程中出现错误则返回相应的错误信息，否则返回nil

func (mp *MailPool) SendMail(subject string, to []string, Cc []string, Bcc []string, message string, annex []string) error {
	// 创建一个新的gomail.Message实例
	m := gomail.NewMessage()

	// 设置发件人信息，使用MailPool中当前可用邮箱账户
	m.SetHeader("From", mp.mailbox[mp.num].name)

	// 设置邮件主题
	m.SetHeader("Subject", subject)

	// 添加收件人
	for _, t := range to {
		m.SetHeader("To", t)
	}

	// 添加抄送人
	for _, c := range Cc {
		m.SetHeader("Cc", c)
	}

	// 添加密送人
	for _, b := range Bcc {
		m.SetHeader("Bcc", b)
	}

	// 设置邮件正文，采用HTML格式
	m.SetBody("text/html", message)

	// 添加附件
	for _, a := range annex {
		m.Attach(a)
	}

	// 创建gomail.Dialer实例，使用当前可用邮箱账户的SMTP服务器配置
	d := gomail.NewDialer(
		mp.mailbox[mp.num].host,
		mp.mailbox[mp.num].port,
		mp.mailbox[mp.num].name,
		mp.mailbox[mp.num].password,
	)

	// 使用DialAndSend方法发送邮件
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	// 若发送成功，返回nil
	return nil
}
