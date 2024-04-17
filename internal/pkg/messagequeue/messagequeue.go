package messagequeue

import (
	"errors"

	"github.com/streadway/amqp"
)

// Connection
// 功能：定义AMQP连接封装结构体，包含指向底层amqp.Connection对象的指针
//
// 字段：
//
//	Connection *amqp.Connection // AMQP连接对象指针，用于与AMQP服务器建立和维护连接
type Connection struct {
	Connection *amqp.Connection // AMQP连接对象指针
}

// MessageQueue
// 功能：定义AMQP消息队列封装结构体，包含指向底层amqp.Channel对象的指针和一个amqp.Queue对象
//
// 字段：
//
//	Channel *amqp.Channel // AMQP通道对象指针，用于执行与消息队列相关的操作
//	Queue amqp.Queue // AMQP队列对象，代表实际的消息队列，包含队列名称、元数据等信息
type MessageQueue struct {
	Channel *amqp.Channel // AMQP通道对象指针
	Queue   amqp.Queue    // AMQP队列对象
}

// NewConnection
// 功能：创建并初始化一个新的Connection实例，通过给定的URL与AMQP服务器建立连接
//
// 参数：
//
//	url string // AMQP服务器URL，字符串类型，用于连接到指定的AMQP服务器
//
// 返回值：
//
//	*Connection // 初始化后的Connection指针，若成功则返回新创建的连接实例，否则返回nil
//	error // 错误信息，若建立连接过程中出现错误则返回相应的错误信息，否则返回nil
func NewConnection(url string) (*Connection, error) {
	// 初始化空的Connection实例
	var c Connection

	// 使用amqp.Dial函数尝试与AMQP服务器建立连接
	var err error

	c.Connection, err = amqp.DialConfig(url, amqp.Config{})

	// 如果建立连接时发生错误，则返回nil的Connection指针和错误信息
	if err != nil {
		return &Connection{}, err
	}

	// 若连接成功，返回初始化后的Connection实例及其指针
	return &c, err
}

// Close
// 功能：关闭Connection实例所持有的AMQP连接
//
// 参数：无
//
// 返回值：
//
//	error // 错误信息，若关闭连接过程中出现错误则返回相应的错误信息，否则返回nil
func (c *Connection) Close() error {

	// 检查Connection.Connection是否为空，若为空则返回错误信息
	if c.Connection == nil {
		return errors.New("Connection.Connection is empty")
	}

	// 调用底层amqp.Connection对象的Close方法关闭连接
	err := c.Connection.Close()

	// 返回关闭连接过程中可能产生的错误
	return err
}

// Close
// 功能：关闭MessageQueue实例所持有的AMQP通道
//
// 参数：无
//
// 返回值：
//
//	error // 错误信息，若关闭通道过程中出现错误则返回相应的错误信息，否则返回nil
func (q MessageQueue) Close() error {
	// 检查MessageQueue.Channel是否为空，若为空则返回错误信息
	if q.Channel == nil {
		return errors.New("MessageQueue.Channel is nil")
	}

	// 调用底层amqp.Channel对象的Close方法关闭通道
	err := q.Channel.Close()

	// 返回关闭通道过程中可能产生的错误
	return err
}

// MessageQueueDeclare
// 功能：在Connection实例所持有的AMQP连接上声明一个新的消息队列，并返回封装后的MessageQueue实例
//
// 参数：
//
//	name string // 队列名称，字符串类型，指定要声明的队列名称
//	durable bool // 是否持久化，布尔类型，若为true，则队列将在服务器重启后保留
//	autoDelete bool // 是否自动删除，布尔类型，若为true，则当最后一个消费者断开连接后，队列将被删除
//	exclusive bool // 是否独占，布尔类型，若为true，则队列仅对当前连接可见，其他连接无法访问
//	noWait bool // 是否等待服务器的响应，布尔类型，若为true，则客户端不等待服务器的响应，立即返回
//	args amqp.Table // 其他参数，amqp.Table类型，包含额外的队列声明参数（如优先级、死信交换器等）
//
// 返回值：
//
//	*MessageQueue // 初始化后的MessageQueue指针，若成功则返回新创建的MessageQueue实例，否则返回nil
//	error // 错误信息，若声明队列过程中出现错误则返回相应的错误信息，否则返回nil
func (c *Connection) MessageQueueDeclare(name string, durable bool, autoDelete bool,
	exclusive bool, noWait bool, args amqp.Table) (*MessageQueue, error) {
	var q MessageQueue
	var err error

	// 检查Connection.Connection是否为空，若为空则返回错误信息
	if c.Connection == nil {
		err = errors.New("Connection.Connection is empty")
		return &MessageQueue{}, err
	}

	// 在连接上打开一个新的AMQP通道
	q.Channel, err = c.Connection.Channel()
	if err != nil {
		return &MessageQueue{}, err
	}

	// 在通道上声明消息队列，使用给定的参数
	q.Queue, err = q.Channel.QueueDeclare(
		name,       // 队列名称
		durable,    // 是否持久化
		autoDelete, // 是否自动删除
		exclusive,  // 是否独占
		noWait,     // 是否等待服务器的响应
		args,       // 其他参数
	)

	// 返回声明好的MessageQueue实例及其指针，以及可能的错误信息
	return &q, err
}

// SendMessage
// 功能：使用MessageQueue实例向指定消息队列发送一条消息
//
// 参数：
//
//	message []byte // 消息内容，字节切片类型，表示要发送的消息数据
//
// 返回值：
//
//	error // 错误信息，若发送消息过程中出现错误则返回相应的错误信息，否则返回nil
func (q *MessageQueue) SendMessage(body []byte) (err error) {
	// 检查MessageQueue.Channel是否为空，若为空则返回错误信息
	if q.Channel == nil {
		err = errors.New("MessageChannel is nil")
		return
	}

	// 使用Channel.Publish方法发布消息到队列
	err = q.Channel.Publish(
		"",           // 交换器名称，这里为空字符串表示直接发送到队列（默认交换器）
		q.Queue.Name, // 路由键名称，使用已声明队列的名称作为路由键
		false,        // 强制发送，布尔类型，若为true且交换器不存在，消息将被丢弃而不返回错误。此处设为false
		false,        // 无法发送的消息将被返回，布尔类型，若为true且无法将消息路由到任何队列，消息将被返回给生产者。此处设为false
		amqp.Publishing{
			ContentType: "application/json", // 内容类型，字符串类型，设置为"application/json"表明消息内容为JSON格式
			Body:        body,               // 消息正文，字节切片类型，包含要发送的实际消息数据
		})

	// 返回发送消息过程中可能产生的错误
	return err
}

// GetMessage
// 功能：使用MessageQueue实例从指定消息队列接收一条消息，并返回一个包含该消息的通道
//
// 参数：无
//
// 返回值：
//
//	message <-chan amqp.Delivery // 消息通道，类型为只读的amqp.Delivery通道，用于接收从队列获取的消息
//	error // 错误信息，若接收消息过程中出现错误则返回相应的错误信息，否则返回nil
func (q *MessageQueue) GetMessage() (message <-chan amqp.Delivery, err error) {
	// 从队列中接收消息，使用Channel.Consume方法设置消费参数
	message, err = q.Channel.Consume(
		q.Queue.Name, // 队列名称，使用已声明队列的名称作为消费的目标队列
		"",           // 消费者标签，空字符串表示不指定消费者标签
		true,         // 自动应答，布尔类型，若为true，则消息被接收后自动确认（acknowledge），否则需要手动确认
		false,        // 独占模式，布尔类型，若为true，则同一时间只有一个消费者可以从队列接收消息。此处设为false
		false,        // 此消费者不用于RabbitMQ的AMQP协议的RPC，布尔类型，此处设为false
		false,        // 不等待服务器的响应，布尔类型，若为true，则客户端不等待服务器的响应，立即返回。此处设为false
		nil,          // 其他参数，此处设为nil，表示不使用额外的消费参数
	)

	// 如果接收消息时发生错误，则返回nil的通道和错误信息
	if err != nil {
		return nil, err
	}

	// 若接收消息成功，返回包含消息的通道
	return message, nil
}
