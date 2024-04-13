package messagequeue

import (
	"errors"

	"github.com/streadway/amqp"
)

type Connection struct {
	Connection *amqp.Connection
}
type MessageQueue struct {
	Channel *amqp.Channel
	Queue   amqp.Queue
}

func NewConnection(url string) (*Connection, error) {
	var c Connection
	var err error
	c.Connection, err = amqp.Dial(url)
	if err != nil {
		return &Connection{}, err
	}
	return &c, err
}
func (c *Connection) Close() error {

	if c.Connection == nil {
		return errors.New("Connection.Connection is empty")
	}
	err := c.Connection.Close()
	return err
}
func (q MessageQueue) Close() error {
	if q.Channel == nil {
		return errors.New("MessageQueue.Channel is nil")
	}
	err := q.Channel.Close()
	return err
}
func (c *Connection) MessageQueueDeclare(name string, durable bool, autoDelete bool,
	exclusive bool, noWait bool, args amqp.Table) (*MessageQueue, error) {
	var q MessageQueue
	var err error
	if c.Connection == nil {
		err = errors.New("Connection.Connection is empty")
		return &MessageQueue{}, err
	}
	q.Channel, err = c.Connection.Channel()
	if err != nil {
		return &MessageQueue{}, err
	}

	q.Queue, err = q.Channel.QueueDeclare(
		name,       // 队列名称
		durable,    // 是否持久化
		autoDelete, // 是否自动删除
		exclusive,  // 是否独占
		noWait,     // 是否等待服务器的响应
		args,       // 其他参数
	)

	return &q, err
}

func (q *MessageQueue) SendMessage(message []byte) (err error) {
	body := []byte(message)

	if q.Channel == nil {
		err = errors.New("MessageChannel is nil")
		return
	}
	// 发布消息到队列
	err = q.Channel.Publish(
		"",           // 交换器名称
		q.Queue.Name, // 路由键名称
		false,        // 强制发送
		false,        // 无法发送的消息将被返回
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	return err
}
func (q *MessageQueue) GetMessage() (message <-chan amqp.Delivery, err error) {
	// 从队列中接收消息
	message, err = q.Channel.Consume(
		q.Queue.Name, // 队列名称
		"",           // 消费者标签，用于区分多个消费者
		true,         // 自动应答，确认收到消息
		false,        // 独占模式
		false,        // 此消费者不用于RabbitMQ的AMQP协议的RPC
		false,        // 不等待服务器的响应
		nil,          // 其他参数
	)
	if err != nil {
		return nil, err
	}
	return

}
