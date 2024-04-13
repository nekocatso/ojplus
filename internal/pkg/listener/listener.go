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

type Listener struct {
	Connection *messagequeue.Connection
	Queue      *messagequeue.MessageQueue
	Ans        int
	Control    chan bool
	Messages   chan []byte
	id         int
	kind       string
	Rcp        *models.Cache
	Rule       map[int]rule.Rule
}

func NewListener(url string, kind string, rcp *models.Cache, id int, Rule map[int]rule.Rule) (*Listener, error) {
	var L Listener
	var err error
	L.Connection, err = messagequeue.NewConnection(url)
	if err != nil {
		return &Listener{}, err
	}
	L.Queue, err = L.Connection.MessageQueueDeclare("queue2", false, false, false, false, nil)
	if err != nil {
		return &Listener{}, err
	}

	L.Ans = 0
	L.Control = make(chan bool)
	L.Messages = make(chan []byte)
	L.kind = kind
	L.Rcp = rcp
	L.id = id
	L.Rule = Rule
	return &L, err
}
func (L Listener) Close() error {

	if L.Queue == nil {
		return errors.New("L.Queue is nil")
	}
	if err := L.Queue.Close(); err != nil {
		return err
	}
	if L.Connection == nil {
		return errors.New("L.Connection is nil")
	}
	if err := L.Connection.Close(); err != nil {
		return err
	}
	close(L.Control)
	return nil
}
func (L Listener) Stop() (err error) {
	L.Control <- true
	return
}
func (L Listener) Listening() (err error) {

	messages, err := L.Queue.GetMessage()
	if err != nil {
		return
	}

	go func(c chan bool, m <-chan amqp.Delivery) {
		for {
			select {
			case msg := <-m:
				if msg.Body == nil {

					return
				}
				L.deal(msg.Body)

			case <-c:
				return
			}
		}
	}(L.Control, messages)
	return
}

func (L Listener) deal(body []byte) {
	log.Println(string(body))
	var res map[string]interface{}
	json.Unmarshal(body, &res)

	L.Rcp.Client.Set(res["correlation_id"].(string), body, 0)
	id, _ := strconv.Atoi(res["correlation_id"].(string))

	if L.Rule[id] != nil {
		// log.Println(string(body))
		L.Rule[id].Scan()
	} else {
		log.Println("rule is nil")
	}
}
