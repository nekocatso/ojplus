package mail

import (
	"reflect"
	"testing"
)

func TestSendMail(t *testing.T) {
	type args struct {
		from *MailBox

		subject string
		to      []string
		Cc      []string
		Bcc     []string
		message string
		annex   []string
	}
	mb := NewMailBox("yangquanmailtest@163.com", "APQJNHKHMXPGRFVO", "smtp.163.com", 25)
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{{name: "success", args: args{from: mb, subject: "sendmailtestsuccess",
		to: []string{"1648806490@qq.com"}, Cc: []string{"yangquanworkmail@163.com"},
		Bcc: []string{"yangquan@ouryun.com.cn"}, message: "this is mail box send mail test ",
		annex: []string{ /* "/home/yq/go2-1/internal/pkg/mail/picture.png" */ }}}} // TODO: Add test cases.

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SendMail(tt.args.from, tt.args.subject, tt.args.to, tt.args.Cc, tt.args.Bcc, tt.args.message, tt.args.annex); (err != nil) != tt.wantErr {
				t.Errorf("SendMail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewMailBox(t *testing.T) {
	type args struct {
		name     string
		password string
		host     string
		port     int
	}
	tests := []struct {
		name string
		args args
		want *MailBox
	}{{name: "success", args: args{name: "", password: "", host: "", port: 1}, want: &MailBox{}}} // TODO: Add test cases.

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMailBox(tt.args.name, tt.args.password, tt.args.host, tt.args.port); !reflect.DeepEqual(got, tt.want) {
				if got == nil {
					t.Errorf("return nil")
				}
			}
		})
	}
}

func TestMaillPool_Mail(t *testing.T) {
	type fields struct {
		mailmap map[string]string
		mailbox *MailBox
	}
	type args struct {
		assets  string
		message string
	}
	mb := NewMailBox("yangquanmailtest@163.com", "APQJNHKHMXPGRFVO", "smtp.163.com", 25)

	tests := []struct {
		name   string
		fields fields
		args   args
	}{{name: "success", fields: fields{mailmap: map[string]string{}, mailbox: mb}, args: args{assets: "asd", message: "this is mappool test"}}} // TODO: Add test cases.

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MaillPool{
				mailmap: tt.fields.mailmap,
				mailbox: tt.fields.mailbox,
			}
			m.Mail(tt.args.assets, tt.args.message, true)
		})
	}
}
