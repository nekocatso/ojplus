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

type ListenerPool struct {
	MaxListener     int
	NumListener     int
	ListenerList    []*listener.Listener
	RedisClientPool *models.Cache
	db              *models.Database
	Rule            map[int]rule.Rule          //用于保存规则对象map,IDtoRule
	cConn           *messagequeue.Connection   //用于和cpp端进行控制信息通信的mq的connection
	cSendQ          *messagequeue.MessageQueue //用于和cpp端发送控制信息的mq的queue
	cGetQ           *messagequeue.MessageQueue //用于和cpp端接收控制信息执行结果的mq的queue
	cGetch          <-chan amqp.Delivery
	mail            *mail.MailBox
}

func NewListenerPool(db *models.Database, RedisClientPool *models.Cache, mail *mail.MailBox) (*ListenerPool, error) {
	var err error
	lp := &ListenerPool{MaxListener: 2, NumListener: 0, ListenerList: []*listener.Listener{},
		RedisClientPool: RedisClientPool, db: db, mail: mail}

	lp.cConn, err = messagequeue.NewConnection("amqp://user:mkjsix7@172.16.0.15:5672/")
	if err != nil {
		return nil, err
	}

	lp.cGetQ, err = lp.cConn.MessageQueueDeclare("queue3", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	lp.cGetch, err = lp.cGetQ.GetMessage()

	if err != nil {
		return nil, err
	}
	lp.cSendQ, err = lp.cConn.MessageQueueDeclare("queue1", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	lp.Rule = make(map[int]rule.Rule)
	//获取所有资产
	a := []models.Asset{}
	err = lp.db.Engine.Where("state <> ?", -1).Find(&a)
	if err != nil {
		return nil, err
	}

	//改变资产状态
	lp.db.Engine.Where("state <> ?", -1).Update(&models.Asset{State: 3})
	//创建规则列表
	for _, i := range a {
		//获取资产对应的规则
		ar := []models.AssetRule{}
		err := lp.db.Engine.Where("asset_id = ?", i.ID).Find(&ar)
		if err != nil {
			return nil, err
		}
		log.Println(i.ID, ar)
		for _, j := range ar {

			//获取ping规则
			p := models.PingInfo{}
			flag, err := lp.db.Engine.Where("id = ?", j.RuleID).Get(&p)
			log.Println(j.RuleID, "ping ", flag)
			if err != nil {
				return nil, err
			}
			if flag {

				err := lp.AddPing(j.ID)
				if err != nil {
					log.Println(err)
				}
			}
			//获取tcp规则
			t := models.TCPInfo{}
			flag, err = lp.db.Engine.Where("id = ?", j.RuleID).Get(&t)
			log.Println(j.RuleID, "tcp ", flag)

			if err != nil {
				return nil, err
			}

			if flag {

				lp.AddTCP(j.ID)
			}

		}

	}
	//log.Println(lp.Rule)
	//兴建监听器
	for i := 0; i < lp.MaxListener; i++ {
		li, err := listener.NewListener("amqp://user:mkjsix7@172.16.0.15:5672/", "respone", lp.RedisClientPool, i, lp.Rule)
		if err != nil {
			return nil, err
		}
		li.Listening()
		lp.ListenerList = append(lp.ListenerList, li)
	}
	fmt.Println("newListenerpool ok")
	return lp, nil
}

func (p *ListenerPool) AddPing(id int) error {
	ar := models.AssetRule{}
	if _, err := p.db.Engine.Where("id = ?", id).Get(&ar); err != nil {
		return err
	}
	a := models.Asset{}
	if _, err := p.db.Engine.Where("id = ?", ar.AssetID).Get(&a); err != nil {
		return err
	}
	r := models.Rule{}
	if _, err := p.db.Engine.Where("id = ?", ar.RuleID).Get(&r); err != nil {
		return err
	}
	m, err := json.Marshal(listener.Request{
		Type:           "request",
		Action:         "ping",
		Target:         []string{a.Address},
		Correlation_id: fmt.Sprintf("%d", id),
		Timestamp:      fmt.Sprintf("%d", time.Now().Unix()),
		Control:        "continuous",
		Config:         []any{"ping", r.Overtime, 1, r.Interval},
	})
	if err != nil {
		return err
	}
	p.cSendQ.SendMessage(m)
	log.Println(string(m))

	for {
		select {
		case c := <-p.cGetch:

			res := listener.RunResult{}
			if err := json.Unmarshal(c.Body, &res); err != nil {
				return err
			}
			fmt.Println(res)
			if res.Status == "success" && res.Corrlation_id == fmt.Sprintf("%d", id) {
				p.Rule[id] = rule.NewPing(id, p.RedisClientPool, p.mail, p.db)
				return nil
			} else {
				return errors.New("add ping failed")
			}
		}
	}

}
func (p *ListenerPool) AddTCP(id int) error {
	ar := models.AssetRule{}
	if _, err := p.db.Engine.Where("id = ?", id).Get(&ar); err != nil {
		return err
	}
	a := models.Asset{}
	if _, err := p.db.Engine.Where("id = ?", ar.AssetID).Get(&a); err != nil {
		return err
	}
	r := models.Rule{}
	if _, err := p.db.Engine.Where("id = ?", ar.RuleID).Get(&r); err != nil {
		return err
	}
	t := models.TCPInfo{}
	if _, err := p.db.Engine.Where("id = ?", ar.RuleID).Get(&t); err != nil {
		return err
	}
	//生成端口地址队列
	add := []string{}
	if t.EnablePorts != "" {
		for _, i := range strings.Split(t.EnablePorts, ",") {
			add = append(add, a.Address+":"+i)
		}
	}
	if t.DisablePorts != "" {
		for _, i := range strings.Split(t.DisablePorts, ",") {
			add = append(add, a.Address+":"+i)
		}
	}

	m, err := json.Marshal(listener.Request{
		Type:           "request",
		Action:         "telnet",
		Target:         add,
		Correlation_id: fmt.Sprintf("%d", id),
		Timestamp:      fmt.Sprintf("%d", time.Now().Unix()),
		Control:        "continuous",
		Config:         []any{"telnet", r.Overtime, 1, r.Interval},
	})
	fmt.Println(string(m))
	if err != nil {
		return err
	}
	p.cSendQ.SendMessage(m)

	if err != nil {
		return err
	}
	for {
		select {
		case c := <-p.cGetch:

			res := listener.RunResult{}
			if err := json.Unmarshal(c.Body, &res); err != nil {
				return err
			}
			fmt.Println(res)
			if res.Status == "success" && res.Corrlation_id == fmt.Sprintf("%d", id) {
				p.Rule[id] = rule.NewTcp(id, p.RedisClientPool, p.mail, p.db)
				return nil
			} else {
				return errors.New("add tcp failed")
			}
		}
	}

}
func (p *ListenerPool) DelPing(id int) error {
	m, err := json.Marshal(listener.Request{
		Type:           "request",
		Action:         "stop_ping",
		Correlation_id: fmt.Sprintf("%d", id),
		Timestamp:      fmt.Sprintf("%d", time.Now().Unix()),
	})
	if err != nil {
		return err
	}
	p.cSendQ.SendMessage(m)

	if err != nil {
		return err
	}
	c := <-p.cGetch
	res := listener.RunResult{}
	if err := json.Unmarshal(c.Body, &res); err != nil {
		return err
	}
	if res.Status == "success" && res.Corrlation_id == fmt.Sprintf("%d", id) {
		delete(p.Rule, id)
		return nil
	} else {
		return errors.New("add ping failed")
	}
}
func (p *ListenerPool) DelTCP(id int) error {
	m, err := json.Marshal(listener.Request{
		Type:           "request",
		Action:         "stop_tcp",
		Correlation_id: fmt.Sprintf("%d", id),
		Timestamp:      fmt.Sprintf("%d", time.Now().Unix()),
	})
	if err != nil {
		return err
	}
	p.cSendQ.SendMessage(m)

	if err != nil {
		return err
	}
	c := <-p.cGetch
	res := listener.RunResult{}
	if err := json.Unmarshal(c.Body, &res); err != nil {
		return err
	}
	if res.Status == "success" && res.Corrlation_id == fmt.Sprintf("%d", id) {
		delete(p.Rule, id)
		return nil
	} else {
		return errors.New("add ping failed")
	}
}
func (p *ListenerPool) UpdatePing(id int) error {
	return p.AddPing(id)
}
func (p *ListenerPool) UpdateTCP(id int) error {
	return p.AddTCP(id)
}
func (p *ListenerPool) Close() {
	m, err := json.Marshal(listener.Request{
		Type:           "request",
		Action:         "stop_all",
		Target:         []string{},
		Correlation_id: fmt.Sprintf("%d", 0),
		Timestamp:      fmt.Sprintf("%d", time.Now().Unix()),
		Config:         []any{},
		Control:        "continuous",
	})
	log.Println(string(m))
	if err != nil {
		log.Println(err)
	}
	p.cSendQ.SendMessage(m)

	select {
	case c := <-p.cGetch:

		res := listener.RunResult{}
		if err := json.Unmarshal(c.Body, &res); err != nil {
			log.Println(err)
		}
		fmt.Println(res)
		if res.Status == "success" {
			log.Println("cpp stop all success")
		} else {
			log.Println("cpp stop all lose")
		}

	}

	p.cGetQ.Close()
	p.cSendQ.Close()
	p.cConn.Close()
	for _, v := range p.ListenerList {
		v.Close()
	}
	log.Println("close all")
}
