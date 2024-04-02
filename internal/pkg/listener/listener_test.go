package listener

import (
	"Alarm/internal/pkg/messagequeue"
	"Alarm/internal/pkg/Cache"
	"testing"
)

func TestNewListener(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    *Listener
		wantErr bool
	}{{name: "success", args: args{url: "amqp://user:mkjsix7@172.16.0.15:5672/"}, want: &Listener{}, wantErr: false},
		{name: "wrongname", args: args{url: "amqp://wrongname:mkjsix7@172.16.0.15:5672/"}, want: &Listener{}, wantErr: true},
		{name: "wrongpassword", args: args{url: "amqp://user:wrongpassword@172.16.0.15:5672/"}, want: &Listener{}, wantErr: true},
		{name: "wrongip", args: args{url: "amqp://user:mkjsix7@wrongip:5672/"}, want: &Listener{}, wantErr: true},
		{name: "wrongport", args: args{url: "amqp://user:mkjsix7@172.16.0.15:wrongport/"}, want: &Listener{}, wantErr: true},
		{name: "wrongpath", args: args{url: "amqp://user:mkjsix7@172.16.0.15:5672wrongpath"}, want: &Listener{}, wantErr: true},
		{name: "withoutname", args: args{url: "amqp://:mkjsix7@172.16.0.15:5672/"}, want: &Listener{}, wantErr: true},
		{name: "withoutpassword", args: args{url: "amqp://user:@172.16.0.15:5672/"}, want: &Listener{}, wantErr: true},
		{name: "withoutip", args: args{url: "amqp://user:mkjsix7@:5672/"}, want: &Listener{}, wantErr: true},
		//{name: "withoutport", args: args{url: "amqp://user:mkjsix7@172.16.0.15:/"}, want:&Listener{}, wantErr: true},
		//{name: "withoutpath", args: args{url: "amqp://user:mkjsix7@172.16.0.15:5672"}, want:&Listener{}, wantErr: true},
		{name: "manyname", args: args{url: "amqp://manyname,user:mkjsix7@172.16.0.15:5672/"}, want: &Listener{}, wantErr: true},
		{name: "manypassword", args: args{url: "amqp://user:manypassword,mkjsix7@172.16.0.15:5672/"}, want: &Listener{}, wantErr: true},
		{name: "manyip", args: args{url: "amqp://user:mkjsix7@manyip,172.16.0.15:5672/"}, want: &Listener{}, wantErr: true},
		{name: "manyport", args: args{url: "amqp://user:mkjsix7@172.16.0.15:5672,manyport/"}, want: &Listener{}, wantErr: true},
		{name: "manypath", args: args{url: "amqp://user:mkjsix7@172.16.0.15:5672/manypath/"}, want: &Listener{}, wantErr: true},

		// TODO: Add test cases.
	}
	rcp := Cache.NewRedisClientPool()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewListener(tt.args.url, "respone", rcp)
			if (err != nil && got != nil) != tt.wantErr {
				t.Errorf("NewListener() error = %v, wantErr %v,got %v", err, tt.wantErr, got)
				return
			}
		})
	}
}

func TestListener_Close(t *testing.T) {
	type fields struct {
		Connection *messagequeue.Connection
		Queue      *messagequeue.MessageQueue
		Ans        int
		Control    chan bool
		Messages   chan []byte
	}
	rcp := Cache.NewRedisClientPool()
	defer rcp.Close()
	l, err := NewListener("amqp://user:mkjsix7@172.16.0.15:5672/", "respone", rcp)
	if err != nil {
		t.Log(err)
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{{name: "success", fields: fields{Connection: l.Connection, Queue: l.Queue, Ans: l.Ans, Control: l.Control, Messages: l.Messages}, wantErr: false},
		{name: "nil Connection", fields: fields{Connection: nil, Queue: l.Queue, Ans: l.Ans, Control: l.Control, Messages: l.Messages}, wantErr: true},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := Listener{
				Connection: tt.fields.Connection,
				Queue:      tt.fields.Queue,
				Ans:        tt.fields.Ans,
				Control:    tt.fields.Control,
				Messages:   tt.fields.Messages,
			}
			if err := L.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Listener.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListener_Stop(t *testing.T) {
	type fields struct {
		Connection *messagequeue.Connection
		Queue      *messagequeue.MessageQueue
		Ans        int
		Control    chan bool
		Messages   chan []byte
	}
	rcp := Cache.NewRedisClientPool()
	defer rcp.Close()
	l, _ := NewListener("amqp://user:mkjsix7@172.16.0.15:5672/", "respone", rcp)
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{{name: "success", fields: fields{Connection: l.Connection, Queue: l.Queue, Ans: l.Ans, Control: l.Control, Messages: l.Messages}, wantErr: false}} // TODO: Add test cases.

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := Listener{
				Connection: tt.fields.Connection,
				Queue:      tt.fields.Queue,
				Ans:        tt.fields.Ans,
				Control:    tt.fields.Control,
				Messages:   tt.fields.Messages,
			}
			l.Listening(func([]byte) {})
			if err := L.Stop(); (err != nil) != tt.wantErr {
				t.Errorf("Listener.Stop() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListener_Listening(t *testing.T) {
	type fields struct {
		Connection *messagequeue.Connection
		Queue      *messagequeue.MessageQueue
		Ans        int
		Control    chan bool
		Messages   chan []byte
	}
	rcp := Cache.NewRedisClientPool()
	defer rcp.Close()
	l, _ := NewListener("amqp://user:mkjsix7@172.16.0.15:5672/", "respone", rcp)
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{{name: "success", fields: fields{Connection: l.Connection, Queue: l.Queue, Ans: l.Ans, Control: l.Control, Messages: l.Messages}, wantErr: false}} // TODO: Add test cases.

	// TODO: Add test cases.

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := Listener{
				Connection: tt.fields.Connection,
				Queue:      tt.fields.Queue,
				Ans:        tt.fields.Ans,
				Control:    tt.fields.Control,
				Messages:   tt.fields.Messages,
			}
			if err := L.Listening(func([]byte) {}); (err != nil) != tt.wantErr {
				t.Errorf("Listener.Listening() error = %v, wantErr %v", err, tt.wantErr)
			}
			l.Stop()
		})
	}
}
