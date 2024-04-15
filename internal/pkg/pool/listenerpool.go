package pool

import (
	"Alarm/internal/pkg/listener"
	"Alarm/internal/pkg/mail"
	"Alarm/internal/pkg/messagequeue"
	"Alarm/internal/pkg/rule"
	"Alarm/internal/web/models"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/streadway/amqp"
)

// ListenerPool
// 功能：定义监听器池结构体，用于管理一组监听器（Listener）及相关资源，如数据库连接、缓存连接、规则映射、与C++端的通信通道等
//
// 字段：
//
//	MaxListener int // 最大监听器数量，整数类型，表示允许同时存在的最大监听器数量上限
//	NumListener int // 当前监听器数量，整数类型，表示当前正在运行的监听器数量
//	ListenerList []*listener.Listener // 监听器列表，监听器指针数组，存储所有已创建的监听器实例
//	RedisClientPool *models.Cache // Redis客户端池，Cache类型的指针，提供对Redis缓存的连接和操作
//	db *models.Database // 数据库连接，Database类型的指针，提供对数据库的连接和操作
//	Rule map[int]rule.Rule // 规则映射表，整数到Rule类型的映射，用于根据ID快速查找对应的规则对象
//	cConn *messagequeue.Connection // 控制信息通信连接，Connection类型的指针，用于与C++端进行控制信息通信的AMQP连接
//	cSendQ *messagequeue.MessageQueue // 控制信息发送队列，MessageQueue类型的指针，用于向C++端发送控制信息的AMQP消息队列
//	cGetQ *messagequeue.MessageQueue // 控制信息接收队列，MessageQueue类型的指针，用于从C++端接收控制信息执行结果的AMQP消息队列
//	cGetch <-chan amqp.Delivery // 控制信息接收通道，只读的amqp.Delivery通道，用于接收从cGetQ队列获取的控制信息执行结果
//	mail *mail.MailPool // 邮件池，MailPool类型的指针，提供邮件发送功能
type ListenerPool struct {
	MaxListener     int                        // 最大监听器数量
	NumListener     int                        // 当前监听器数量
	ListenerList    []*listener.Listener       // 监听器列表
	RedisClientPool *models.Cache              // Redis客户端池
	db              *models.Database           // 数据库连接，Database类型的指针
	Rule            map[int]rule.Rule          // 用于保存规则对象的映射表，ID到Rule对象的映射
	cConn           *messagequeue.Connection   // 用于和C++端进行控制信息通信的AMQP连接
	cSendQ          *messagequeue.MessageQueue // 用于和C++端发送控制信息的AMQP消息队列
	cGetQ           *messagequeue.MessageQueue // 用于和C++端接收控制信息执行结果的AMQP消息队列
	cGetch          <-chan amqp.Delivery       // 用于接收从cGetQ队列获取的控制信息执行结果的通道
	mail            *mail.MailPool             // 邮件池，提供邮件发送功能
}

// NewListenerPool
// 功能：创建并初始化一个新的监听器池实例，包括数据库连接、缓存连接、邮件池、与C++端的通信通道及监听器列表等资源
//
// 参数：
//
//	db *models.Database // 数据库连接，Database类型的指针，用于与数据库进行交互
//	RedisClientPool *models.Cache // Redis客户端池，Cache类型的指针，用于与Redis缓存进行交互
//	mail *mail.MailPool // 邮件池，MailPool类型的指针，用于发送邮件通知
//	url string // AMQP连接URL，字符串类型，用于连接到AMQP消息代理服务器
//
// 返回值：
//
//	*ListenerPool // 新建的监听器池实例，类型为ListenerPool的指针，若成功创建则返回该实例，否则返回nil
//	error // 错误信息，若初始化过程中出现错误则返回相应的错误信息，否则返回nil
func NewListenerPool(db *models.Database, RedisClientPool *models.Cache, mail *mail.MailPool, url string) (*ListenerPool, error) {
	if db == nil {
		return nil, errors.New("db is nil") // 检查数据库连接是否为空，若为空则返回错误信息
	}
	if RedisClientPool == nil {
		return nil, errors.New("RedisClientPool is nil") // 检查Redis客户端池是否为空，若为空则返回错误信息
	}
	if mail == nil {
		return nil, errors.New("mail is nil") // 检查邮件池是否为空，若为空则返回错误信息
	}

	var err error
	lp := &ListenerPool{MaxListener: 2, NumListener: 0, ListenerList: []*listener.Listener{},
		RedisClientPool: RedisClientPool, db: db, mail: mail} // 创建并初始化监听器池实例

	// 建立与C++端的AMQP连接
	lp.cConn, err = messagequeue.NewConnection(url)
	if err != nil {
		return nil, err // 若建立连接失败，则返回错误信息
	}

	// 声明用于接收控制信息执行结果的AMQP消息队列（queue3）并获取其接收通道
	lp.cGetQ, err = lp.cConn.MessageQueueDeclare("queue3", false, false, false, false, nil)
	if err != nil {
		return nil, err // 若声明队列或获取通道失败，则返回错误信息
	}
	lp.cGetch, err = lp.cGetQ.GetMessage()
	if err != nil {
		return nil, err // 若获取通道失败，则返回错误信息
	}

	// 声明用于发送控制信息的AMQP消息队列（queue1）
	lp.cSendQ, err = lp.cConn.MessageQueueDeclare("queue1", false, false, false, false, nil)
	if err != nil {
		return nil, err // 若声明队列失败，则返回错误信息
	}

	// 初始化规则映射表
	lp.Rule = make(map[int]rule.Rule)

	// 从数据库中查询所有非未监控状态的资产（Asset）
	a := []models.Asset{}
	err = lp.db.Engine.Where("state <> ?", -1).Find(&a)
	if err != nil {
		return nil, err // 若查询失败，则返回错误信息
	}

	// 更新所有非未监控状态的资产至状态3
	lp.db.Engine.Where("state <> ?", -1).Update(&models.Asset{State: 3})

	// 遍历查询到的资产，为其关联的规则创建监听器
	for _, i := range a {
		ar := []models.AssetRule{} // 查询与当前资产关联的规则
		err := lp.db.Engine.Where("asset_id = ?", i.ID).Find(&ar)
		if err != nil {
			return nil, err // 若查询失败，则返回错误信息
		}
		log.Println(ar)
		// 遍历资产关联的规则，根据规则类型（Ping或TCP）创建相应的监听器
		for _, j := range ar {
			p := models.PingInfo{} // 查询Ping规则详细信息
			flag, err := lp.db.Engine.Where("id = ?", j.RuleID).Get(&p)
			log.Println(j.RuleID, "ping ", flag)
			if err != nil {
				return nil, err // 若查询失败，则返回错误信息
			}
			if flag { // 若存在Ping规则，则添加Ping监听器
				err := lp.AddPing(j.ID)
				if err != nil {
					log.Println(err)
				}
			}

			t := models.TCPInfo{} // 查询TCP规则详细信息
			flag, err = lp.db.Engine.Where("id = ?", j.RuleID).Get(&t)
			log.Println(j.RuleID, "tcp ", flag)
			if err != nil {

				return nil, err // 若查询失败，则返回错误信息
			}
			if flag { // 若存在TCP规则，则添加TCP监听器
				lp.AddTCP(j.ID)
			}
		}
	}
	log.Println("rule built end")
	// 根据最大监听器数量创建并启动监听器实例，添加到监听器列表
	for i := 0; i < lp.MaxListener; i++ {
		li, err := listener.NewListener("amqp://user:mkjsix7@172.16.0.15:5672/", lp.RedisClientPool, i+1, lp.Rule)
		if err != nil {
			return nil, err // 若创建监听器失败，则返回错误信息
		}
		li.Listening()
		lp.ListenerList = append(lp.ListenerList, li)
	}

	log.Println("newListenerpool ok")
	return lp, nil // 初始化成功，返回新建的监听器池实例和nil错误信息
}

// AddPing
// 功能：为监听器池添加一个Ping类型的监听任务，包括从数据库获取相关资产、规则信息，构造请求消息发送至C++端，并等待接收执行结果
//
// 参数：
//
//	id int // 监听任务ID，整数类型，标识要添加的Ping监听任务，应为资产规则映射表中的主键
//
// 返回值：
//
//	error // 错误信息，若添加过程出现错误则返回相应的错误信息，否则返回nil
func (p *ListenerPool) AddPing(id int) error {
	// 从数据库中获取与监听任务ID关联的AssetRule记录
	ar := models.AssetRule{}
	if _, err := p.db.Engine.Where("id = ?", id).Get(&ar); err != nil {
		return err // 获取AssetRule失败，则返回错误信息
	}

	// 从数据库中获取与AssetRule关联的Asset记录
	a := models.Asset{}
	if _, err := p.db.Engine.Where("id = ?", ar.AssetID).Get(&a); err != nil {
		return err // 获取Asset失败，则返回错误信息
	}

	// 从数据库中获取与AssetRule关联的Rule记录
	r := models.Rule{}
	if _, err := p.db.Engine.Where("id = ?", ar.RuleID).Get(&r); err != nil {
		return err // 获取Rule失败，则返回错误信息
	}

	// 构造并向C++端发送Ping请求消息
	m, err := json.Marshal(listener.Request{
		Type:           "request",
		Action:         "ping",
		Target:         []string{a.Address},                      // 目标地址，取自Asset的Address字段
		Correlation_id: fmt.Sprintf("%d", id),                    // 关联ID，使用监听任务ID
		Timestamp:      fmt.Sprintf("%d", time.Now().Unix()),     // 时间戳，当前时间的Unix时间戳
		Control:        "continuous",                             // 控制类型，固定为"continuous"
		Config:         []any{"ping", r.Overtime, 1, r.Interval}, // 配置参数，包括Ping命令、超时时间、重试次数、间隔时间
	})
	if err != nil {
		return err // 构造或发送消息失败，则返回错误信息
	}
	p.cSendQ.SendMessage(m) // 发送消息至C++端
	log.Println(string(m))  // 打印发送的消息内容

	// 循环等待接收C++端返回的执行结果
	for {
		select {
		case c := <-p.cGetch: // 从控制信息接收通道接收消息

			// 反序列化接收到的消息为RunResult结构体
			res := listener.RunResult{}
			if err := json.Unmarshal(c.Body, &res); err != nil {
				return err // 反序列化失败，则返回错误信息
			}
			fmt.Println(res) // 打印接收到的执行结果

			// 检查执行结果状态和关联ID，若成功且匹配，则添加Ping监听器至规则映射表并返回nil
			if res.Status == "success" && res.Corrlation_id == fmt.Sprintf("%d", id) {
				p.Rule[id] = rule.NewPing(id, p.RedisClientPool, p.mail, p.db)
				return nil
			} else {
				return errors.New("add ping failed") // 若执行结果状态不是成功或关联ID不匹配，则返回添加失败的错误信息
			}
		}
	}
}

// AddTCP
// 功能：为监听器池添加一个TCP类型的监听任务，包括从数据库获取相关资产、规则、TCP信息，构造请求消息发送至C++端，并等待接收执行结果
//
// 参数：
//
//	id int // 监听任务ID，整数类型，标识要添加的TCP监听任务
//
// 返回值：
//
//	error // 错误信息，若添加过程出现错误则返回相应的错误信息，否则返回nil
func (p *ListenerPool) AddTCP(id int) error {
	// 从数据库中获取与监听任务ID关联的AssetRule记录
	ar := models.AssetRule{}
	if _, err := p.db.Engine.Where("id = ?", id).Get(&ar); err != nil {
		return err // 获取AssetRule失败，则返回错误信息
	}

	// 从数据库中获取与AssetRule关联的Asset记录
	a := models.Asset{}
	if _, err := p.db.Engine.Where("id = ?", ar.AssetID).Get(&a); err != nil {
		return err // 获取Asset失败，则返回错误信息
	}

	// 从数据库中获取与AssetRule关联的Rule记录
	r := models.Rule{}
	if _, err := p.db.Engine.Where("id = ?", ar.RuleID).Get(&r); err != nil {
		return err // 获取Rule失败，则返回错误信息
	}

	// 从数据库中获取与AssetRule关联的TCPInfo记录
	t := models.TCPInfo{}
	if _, err := p.db.Engine.Where("id = ?", ar.RuleID).Get(&t); err != nil {
		return err // 获取TCPInfo失败，则返回错误信息
	}

	// 生成目标端口地址列表
	add := []string{}
	if t.EnablePorts != "" {
		for _, i := range strings.Split(t.EnablePorts, ",") {
			add = append(add, a.Address+":"+i) // 将资产地址与启用端口拼接为完整的目标地址
		}
	}
	if t.DisablePorts != "" {
		for _, i := range strings.Split(t.DisablePorts, ",") {
			add = append(add, a.Address+":"+i) // 将资产地址与禁用端口拼接为完整的目标地址
		}
	}

	// 构造TCP请求消息
	m, err := json.Marshal(listener.Request{
		Type:           "request",
		Action:         "telnet",                                   // 行动类型，固定为"telnet"
		Target:         add,                                        // 目标地址列表，包含启用端口和禁用端口的目标地址
		Correlation_id: fmt.Sprintf("%d", id),                      // 关联ID，使用监听任务ID
		Timestamp:      fmt.Sprintf("%d", time.Now().Unix()),       // 时间戳，当前时间的Unix时间戳
		Control:        "continuous",                               // 控制类型，固定为"continuous"
		Config:         []any{"telnet", r.Overtime, 1, r.Interval}, // 配置参数，包括Telnet命令、超时时间、重试次数、间隔时间
	})
	fmt.Println(string(m)) // 打印发送的消息内容
	if err != nil {
		return err // 构造失败，则返回错误信息
	}
	err = p.cSendQ.SendMessage(m) // 发送消息至C++端
	if err != nil {
		return err // 发送失败，则返回错误信息
	}
	// 循环等待接收C++端返回的执行结果
	for {
		select {
		case c := <-p.cGetch: // 从控制信息接收通道接收消息

			// 反序列化接收到的消息为RunResult结构体
			res := listener.RunResult{}
			if err := json.Unmarshal(c.Body, &res); err != nil {
				return err // 反序列化失败，则返回错误信息
			}
			fmt.Println(res) // 打印接收到的执行结果

			// 检查执行结果状态和关联ID，若成功且匹配，则添加TCP监听器至规则映射表并返回nil
			if res.Status == "success" && res.Corrlation_id == fmt.Sprintf("%d", id) {
				p.Rule[id] = rule.NewTcp(id, p.RedisClientPool, p.mail, p.db)
				return nil
			} else {
				return errors.New("add tcp failed") // 若执行结果状态不是成功或关联ID不匹配，则返回添加失败的错误信息
			}
		}
	}
}

// DelPing
// 功能：从监听器池中移除指定ID的Ping监听任务，包括构造停止请求消息发送至C++端，并等待接收执行结果
//
// 参数：
//
//	id int // 监听任务ID，整数类型，标识要移除的Ping监听任务
//
// 返回值：
//
//	error // 错误信息，若移除过程出现错误则返回相应的错误信息，否则返回nil
func (p *ListenerPool) DelPing(id int) error {
	// 构造并向C++端发送停止Ping请求消息
	m, err := json.Marshal(listener.Request{
		Type:           "request",
		Action:         "stop_ping",                          // 行动类型，固定为"stop_ping"
		Correlation_id: fmt.Sprintf("%d", id),                // 关联ID，使用监听任务ID
		Timestamp:      fmt.Sprintf("%d", time.Now().Unix()), // 时间戳，当前时间的Unix时间戳
	})
	if err != nil {
		return err // 构造或发送消息失败，则返回错误信息
	}
	p.cSendQ.SendMessage(m) // 发送消息至C++端

	// 接收C++端返回的执行结果
	c := <-p.cGetch // 从控制信息接收通道接收消息

	// 反序列化接收到的消息为RunResult结构体
	res := listener.RunResult{}
	if err := json.Unmarshal(c.Body, &res); err != nil {
		return err // 反序列化失败，则返回错误信息
	}

	// 检查执行结果状态和关联ID，若成功且匹配，则更新Ping监听器状态并从规则映射表中移除，然后返回nil
	if res.Status == "success" && res.Corrlation_id == fmt.Sprintf("%d", id) {
		p.Rule[id].Update() // 更新Ping监听器状态
		delete(p.Rule, id)  // 从规则映射表中移除指定ID的Ping监听器
		return nil
	} else {
		return errors.New("del ping failed") // 若执行结果状态不是成功或关联ID不匹配，则返回移除失败的错误信息
	}
}

// DelTCP
// 功能：从监听器池中移除指定ID的TCP监听任务，包括构造停止请求消息发送至C++端，并等待接收执行结果
//
// 参数：
//
//	id int // 监听任务ID，整数类型，标识要移除的TCP监听任务
//
// 返回值：
//
//	error // 错误信息，若移除过程出现错误则返回相应的错误信息，否则返回nil
func (p *ListenerPool) DelTCP(id int) error {
	// 构造并向C++端发送停止TCP请求消息
	m, err := json.Marshal(listener.Request{
		Type:           "request",
		Action:         "stop_tcp",                           // 行动类型，固定为"stop_tcp"
		Correlation_id: fmt.Sprintf("%d", id),                // 关联ID，使用监听任务ID
		Timestamp:      fmt.Sprintf("%d", time.Now().Unix()), // 时间戳，当前时间的Unix时间戳
	})
	if err != nil {
		return err // 构造或发送消息失败，则返回错误信息
	}
	p.cSendQ.SendMessage(m) // 发送消息至C++端

	// 接收C++端返回的执行结果
	c := <-p.cGetch // 从控制信息接收通道接收消息

	// 反序列化接收到的消息为RunResult结构体
	res := listener.RunResult{}
	if err := json.Unmarshal(c.Body, &res); err != nil {
		return err // 反序列化失败，则返回错误信息
	}

	// 检查执行结果状态和关联ID，若成功且匹配，则更新TCP监听器状态并从规则映射表中移除，然后返回nil
	if res.Status == "success" && res.Corrlation_id == fmt.Sprintf("%d", id) {
		p.Rule[id].Update() // 更新TCP监听器状态
		delete(p.Rule, id)  // 从规则映射表中移除指定ID的TCP监听器
		return nil
	} else {
		return errors.New("del tcp failed") // 若执行结果状态不是成功或关联ID不匹配，则返回移除失败的错误信息
	}
}

// UpdatePing
// 功能：更新指定ID的Ping监听任务，包括调用监听器的Update方法更新其状态，然后重新添加该Ping监听任务到监听器池
//
// 参数：
//
//	id int // 监听任务ID，整数类型，标识要更新的Ping监听任务
//
// 返回值：
//
//	error // 错误信息，若更新或重新添加过程出现错误则返回相应的错误信息，否则返回nil
func (p *ListenerPool) UpdatePing(id int) error {
	// 调用Ping监听器的Update方法更新其状态
	p.Rule[id].Update()

	// 重新添加更新后的Ping监听任务到监听器池
	return p.AddPing(id) // 返回AddPing方法的结果作为本方法的返回值
}

// UpdateTCP
// 功能：更新指定ID的TCP监听任务，包括调用监听器的Update方法更新其状态，然后重新添加该TCP监听任务到监听器池
//
// 参数：
//
//	id int // 监听任务ID，整数类型，标识要更新的TCP监听任务
//
// 返回值：
//
//	error // 错误信息，若更新或重新添加过程出现错误则返回相应的错误信息，否则返回nil
func (p *ListenerPool) UpdateTCP(id int) error {
	// 调用TCP监听器的Update方法更新其状态
	p.Rule[id].Update()

	// 重新添加更新后的TCP监听任务到监听器池
	return p.AddTCP(id) // 返回AddTCP方法的结果作为本方法的返回值
}

// Close
// 功能：关闭监听器池，包括发送停止所有监听任务的请求至C++端，等待接收执行结果，关闭相关通道及连接，并关闭所有监听器
//
// 参数：无
//
// 返回值：无

func (p *ListenerPool) Close() {
	// 构造并向C++端发送停止所有监听任务的请求消息
	m, err := json.Marshal(listener.Request{
		Type:           "request",
		Action:         "stop_all",                           // 行动类型，固定为"stop_all"
		Target:         []string{},                           // 空目标列表，因停止所有任务无需特定目标
		Correlation_id: fmt.Sprintf("%d", 0),                 // 关联ID，使用0表示停止所有任务
		Timestamp:      fmt.Sprintf("%d", time.Now().Unix()), // 时间戳，当前时间的Unix时间戳
		Config:         []any{},                              // 空配置参数列表，因停止所有任务无需特定配置
		Control:        "continuous",                         // 控制类型，固定为"continuous"
	})
	log.Println(string(m)) // 打印发送的消息内容
	if err != nil {
		log.Println(err) // 构造或发送消息失败，打印错误信息
	}
	err = p.cSendQ.SendMessage(m) // 发送消息至C++端
	if err != nil {
		log.Println(err)
	}
	// 接收C++端返回的执行结果
	c := <-p.cGetch // 从控制信息接收通道接收消息

	// 反序列化接收到的消息为RunResult结构体
	res := listener.RunResult{}
	if err := json.Unmarshal(c.Body, &res); err != nil {
		log.Println(err) // 反序列化失败，打印错误信息
	}
	fmt.Println(res) // 打印接收到的执行结果

	// 根据执行结果状态输出相应日志信息
	if res.Status == "success" {
		log.Println("cpp stop all success") // 停止所有任务成功
	} else {
		log.Println("cpp stop all lose") // 停止所有任务失败
	}

	// 关闭相关通道及连接
	p.cGetQ.Close()  // 关闭控制信息接收队列
	p.cSendQ.Close() // 关闭控制信息发送队列
	p.cConn.Close()  // 关闭与C++端的连接

	// 关闭所有监听器
	for _, v := range p.ListenerList {
		v.Close() // 关闭单个监听器
	}
	log.Println("close all") // 输出关闭所有资源的日志信息
}
