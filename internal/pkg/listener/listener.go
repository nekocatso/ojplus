package listener

import (
	"Alarm/internal/pkg/messagequeue"
	"Alarm/internal/pkg/rule"
	"Alarm/internal/web/models"
	"encoding/json"
	"errors"
	"log"
	"strconv"

	"github.com/streadway/amqp"
)

// 提供监听功能的对象
type Listener struct {
	Connection *messagequeue.Connection   /* 提供和消息队列的连接 */
	Queue      *messagequeue.MessageQueue /* 提供通信用消息队列 */
	Ans        int                        /* 收件数计数器 */
	Control    chan bool                  /* 控制队列用以关闭监听进程 */
	Messages   chan []byte                /* 获取queue中消息的通道 */
	Id         int                        /* 每个监听者的标识id */
	Cache      *models.Cache              /*  用于和redis通信 */
	Rule       map[int]rule.Rule          /* 用于与规则对象通讯 */
}

// NewListener 创建一个新的Listener实例。
//
// 参数:
//
//	url: 用于建立连接的URL，不能为空。
//	rcp: 指向Cache实例的指针，用于存储和检索数据，不能为nil。
//	Id: 唯一标识符，必须大于0。
//	Rule: 规则映射，键为int类型，值为rule.Rule类型，用于消息处理规则，不能为nil。
//
// 返回值:
//
//	*Listener: 成功时返回初始化好的Listener指针。
//	error: 创建过程中遇到错误时返回error。
func NewListener(url string, rcp *models.Cache, Id int, Rule map[int]rule.Rule) (*Listener, error) {
	// 参数检查
	if url == "" {
		return nil, errors.New("url is empty")
	}
	if rcp == nil {
		return nil, errors.New("rcp is nil")
	}
	if Id == 0 {
		return nil, errors.New("id is 0")
	}
	if Rule == nil {
		return nil, errors.New("rule is nil")
	}

	var L Listener
	var err error
	// 建立连接
	L.Connection, err = messagequeue.NewConnection(url)
	if err != nil {
		return nil, err
	}
	// 创建消息队列
	L.Queue, err = L.Connection.MessageQueueDeclare("queue2", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	// 初始化参数
	L.Ans = 0
	L.Control = make(chan bool)
	L.Messages = make(chan []byte)
	L.Cache = rcp
	L.Id = Id
	L.Rule = Rule
	return &L, err
}

// Close 关闭Listener的消息队列和连接。
// 返回可能发生的任何错误。
func (L Listener) Close() error {
	// 检查消息队列是否为nil并尝试关闭
	if L.Queue == nil {
		return errors.New("L.Queue is nil")
	}
	if err := L.Queue.Close(); err != nil {
		return err
	}

	// 检查连接是否为nil并尝试关闭
	if L.Connection == nil {
		return errors.New("L.Connection is nil")
	}
	if err := L.Connection.Close(); err != nil {
		return err
	}

	// 关闭控制通道
	close(L.Control)
	return nil
}

// Stop停止监听器的操作。
// 此函数向Listener的Control通道发送一个true值，以指示需要停止监听。
//
// 参数:
// L Listener - 实现了Listener接口的监听器对象。
//
// 返回值:
// err error - 停止操作过程中遇到的任何错误。如果操作成功，返回nil。
func (L Listener) Stop() (err error) {
	L.Control <- true // 向Control通道发送停止信号
	return
}

// Listening 方法用于启动消息监听。
// 该方法首先从队列中获取消息，然后使用goroutine处理这些消息。
// 如果成功启动监听，方法将无返回值；如果遇到错误，则返回错误信息。
func (L Listener) Listening() (err error) {
	// 尝试从队列获取消息，如果出错则直接返回错误
	messages, err := L.Queue.GetMessage()
	if err != nil {
		return
	}

	// 使用goroutine启动一个监听线程，处理从队列中获取的消息
	go func(c chan bool, m <-chan amqp.Delivery) {
		for {
			// 通过select语句监听两个通道：消息通道和控制通道
			select {
			case msg := <-m:
				L.Ans++ // 每处理一条消息，计数器增加
				if msg.Body == nil {
					// 如果消息体为空，则退出当前循环
					return
				}
				L.deal(msg.Body) // 处理消息体

			case <-c:
				// 如果接收到控制通道的消息，则退出监听
				return
			}
		}
	}(L.Control, messages)
	return
}

// deal
// 功能：处理监听器接收到的消息体。
// 参数：
//   1. body: 接收的消息体字节数组，用于存放待处理的数据。
// 返回：无

func (L Listener) deal(body []byte) {
	// 打印接收到的消息体内容
	log.Println(string(body))

	var res map[string]interface{} // 定义一个map用于存储解析后的JSON数据

	// 将JSON数据解析到res变量中
	json.Unmarshal(body, &res)

	// 使用correlation_id作为键，将原始消息体存储到缓存中
	L.Cache.Client.Set(res["correlation_id"].(string), body, 0)

	// 将correlation_id转换为int类型
	id, _ := strconv.Atoi(res["correlation_id"].(string))
	//log.Println(id)
	// 检查是否存在对应id的规则，如果存在，则执行扫描操作
	if L.Rule[id] != nil {

		err := L.Rule[id].Scan()
		if err != nil {
			log.Println(err)
		}
	} else {
		// 如果不存在对应id的规则，打印日志信息
		log.Println("rule is nil")
	}
}
