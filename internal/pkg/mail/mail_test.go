package mail

import (
	"reflect"
	"testing"
)

func TestSendMail(t *testing.T) {
	type args struct {
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
	}{{name: "success", args: args{subject: "【告警】xx资产-【规则】-异常触发",
		to: []string{"1648806490@qq.com"}, Cc: []string{},
		Bcc: []string{}, message: `告警类型：PING检测/TCP端口探测<br>
		告警节点：异常触发<br>
		告警资产：资产1<br>
		资产地址：192.168.1.1<br>
		检测规则：规则1<br>
		告警内容：
	  	若为PING检测：<br>
		&nbsp&nbsp&nbsp&nbsp响应时间大于等于X ms（附该次响应时间），丢包率大于等于X %（x+1%）<br>
		告警时间：2024-03-18  09:30:30<br>
		该资产在此规则监控下触发异常，请尽快处理！`,
		annex: []string{ /* "/home/yq/go2-1/internal/pkg/mail/picture.png" */ }}}} // TODO: Add test cases.

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := mb.SendMail(tt.args.subject, tt.args.to, tt.args.Cc, tt.args.Bcc, tt.args.message, tt.args.annex); (err != nil) != tt.wantErr {
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
			 
			
			
		})
	}
}
