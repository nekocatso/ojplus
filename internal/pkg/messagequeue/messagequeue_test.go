package messagequeue

import (
	"log"
	"testing"

	"github.com/streadway/amqp"
)

func TestConnection_close(t *testing.T) {
	type fields struct {
		Connection *amqp.Connection
	}
	c, _ := amqp.Dial("amqp://user:mkjsix7@172.16.0.15:5672/")
	defer c.Close()
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{{name: "success", fields: fields{Connection: c}, wantErr: false},
		{name: "empty", fields: fields{Connection: nil}, wantErr: true},

		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Connection{
				Connection: tt.fields.Connection,
			}
			if err := c.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Connection.close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_MessageQueue_close(t *testing.T) {
	type fields struct {
		Channel *amqp.Channel
		Queue   amqp.Queue
	}
	conn, err := amqp.Dial("amqp://user:mkjsix7@172.16.0.15:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer conn.Close()

	// 打开一个通道
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}
	defer ch.Close()

	// 声明要从中接收消息的队列
	q, err := ch.QueueDeclare(
		"hello", // 队列名称
		false,   // 是否持久化
		false,   // 是否自动删除
		false,   // 是否独占
		false,   // 是否等待服务器的响应
		nil,     // 其他参数
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %s", err)
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{{name: "success", fields: fields{Channel: ch, Queue: q}, wantErr: false},
		{name: "empty ch", fields: fields{Channel: nil, Queue: q}, wantErr: true},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := MessageQueue{
				Channel: tt.fields.Channel,
				Queue:   tt.fields.Queue,
			}
			if err := q.Close(); (err != nil) != tt.wantErr {
				t.Errorf("MessageQueue.close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
func TestNewConnection(t *testing.T) {
	type args struct {
		amqp string
	}
	tests := []struct {
		name    string
		args    args
		wantC   Connection
		wantErr bool
	}{{name: "success", args: args{amqp: "amqp://user:mkjsix7@172.16.0.15:5672/"}, wantC: Connection{}, wantErr: false},
		{name: "wrongname", args: args{amqp: "amqp://wrongname:mkjsix7@172.16.0.15:5672/"}, wantC: Connection{}, wantErr: true},
		{name: "wrongpassword", args: args{amqp: "amqp://user:wrongpassword@172.16.0.15:5672/"}, wantC: Connection{}, wantErr: true},
		{name: "wrongip", args: args{amqp: "amqp://user:mkjsix7@wrongip:5672/"}, wantC: Connection{}, wantErr: true},
		{name: "wrongport", args: args{amqp: "amqp://user:mkjsix7@172.16.0.15:wrongport/"}, wantC: Connection{}, wantErr: true},
		{name: "wrongpath", args: args{amqp: "amqp://user:mkjsix7@172.16.0.15:5672wrongpath"}, wantC: Connection{}, wantErr: true},
		{name: "withoutname", args: args{amqp: "amqp://:mkjsix7@172.16.0.15:5672/"}, wantC: Connection{}, wantErr: true},
		{name: "withoutpassword", args: args{amqp: "amqp://user:@172.16.0.15:5672/"}, wantC: Connection{}, wantErr: true},
		{name: "withoutip", args: args{amqp: "amqp://user:mkjsix7@:5672/"}, wantC: Connection{}, wantErr: true},
		//{name: "withoutport", args: args{amqp: "amqp://user:mkjsix7@172.16.0.15:/"}, wantC: Connection{}, wantErr: true},
		//{name: "withoutpath", args: args{amqp: "amqp://user:mkjsix7@172.16.0.15:5672"}, wantC: Connection{}, wantErr: true},
		{name: "manyname", args: args{amqp: "amqp://manyname,user:mkjsix7@172.16.0.15:5672/"}, wantC: Connection{}, wantErr: true},
		{name: "manypassword", args: args{amqp: "amqp://user:manypassword,mkjsix7@172.16.0.15:5672/"}, wantC: Connection{}, wantErr: true},
		{name: "manyip", args: args{amqp: "amqp://user:mkjsix7@manyip,172.16.0.15:5672/"}, wantC: Connection{}, wantErr: true},
		{name: "manyport", args: args{amqp: "amqp://user:mkjsix7@172.16.0.15:5672,manyport/"}, wantC: Connection{}, wantErr: true},
		{name: "manypath", args: args{amqp: "amqp://user:mkjsix7@172.16.0.15:5672/manypath/"}, wantC: Connection{}, wantErr: true},

		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewConnection(tt.args.amqp)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewConnection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if c.Connection != nil {
				c.Close()
			}

		})
	}
}

func TestConnection_MessageQueueDeclare(t *testing.T) {
	type fields struct {
		Connection *amqp.Connection
	}
	type args struct {
		name       string
		durable    bool
		autoDelete bool
		exclusive  bool
		noWait     bool
		args       amqp.Table
	}
	ch, _ := amqp.Dial("amqp://user:mkjsix7@172.16.0.15:5672/")
	defer ch.Close()
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantQ   MessageQueue
		wantErr bool
	}{{name: "success", fields: fields{Connection: ch}, args: args{name: "test"}, wantQ: MessageQueue{}, wantErr: false},
		{name: "emptyconnection", fields: fields{Connection: nil}, args: args{}, wantQ: MessageQueue{}, wantErr: true},
		{name: "emptyname", fields: fields{Connection: ch}, args: args{name: ""}, wantQ: MessageQueue{}, wantErr: false},

		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Connection{
				Connection: tt.fields.Connection,
			}
			gotQ, err := c.MessageQueueDeclare(tt.args.name, tt.args.durable, tt.args.autoDelete, tt.args.exclusive, tt.args.noWait, tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Connection.MessageQueueDeclare() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotQ.Channel != nil {
				gotQ.Close()
			}

		})
	}
}
func Test_MessageQueue_SendMessage(t *testing.T) {
	type fields struct {
		Channel *amqp.Channel
		Queue   amqp.Queue
	}
	type args struct {
		message []byte
	}
	conn, err := amqp.Dial("amqp://user:mkjsix7@172.16.0.15:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer conn.Close()

	// 打开一个通道
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}
	defer ch.Close()

	// 声明要从中接收消息的队列
	q, err := ch.QueueDeclare(
		"hello", // 队列名称
		false,   // 是否持久化
		false,   // 是否自动删除
		false,   // 是否独占
		false,   // 是否等待服务器的响应
		nil,     // 其他参数
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %s", err)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "success", fields: fields{Channel: ch, Queue: q}, args: args{message: []byte{}}, wantErr: false},
		{name: "empty ch", fields: fields{Channel: nil, Queue: q}, args: args{message: []byte{}}, wantErr: true},
		{name: "empty q", fields: fields{Channel: ch, Queue: amqp.Queue{}}, args: args{message: []byte{}}, wantErr: false},

		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &MessageQueue{
				Channel: tt.fields.Channel,
				Queue:   tt.fields.Queue,
			}
			if err := q.SendMessage(tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("MessageQueue.SendMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMessageQueue_GetMessage(t *testing.T) {
	type fields struct {
		Channel *amqp.Channel
		Queue   amqp.Queue
	}
	type args struct {
		control chan bool
		message chan []byte
	}
	conn, err := amqp.Dial("amqp://user:mkjsix7@172.16.0.15:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer conn.Close()

	// 打开一个通道
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}
	defer ch.Close()

	// 声明要从中接收消息的队列
	q, err := ch.QueueDeclare(
		"hello", // 队列名称
		false,   // 是否持久化
		false,   // 是否自动删除
		false,   // 是否独占
		false,   // 是否等待服务器的响应
		nil,     // 其他参数
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %s", err)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{{name: "success", fields: fields{Channel: ch, Queue: q}, args: args{}, wantErr: false}} // TODO: Add test cases.

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &MessageQueue{
				Channel: tt.fields.Channel,
				Queue:   tt.fields.Queue,
			}
			if _, err := q.GetMessage(); (err != nil) != tt.wantErr {
				t.Errorf("MessageQueue.GetMessage() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}
