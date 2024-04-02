package listener

import (
	"Alarm/internal/pkg/messagequeue"
	"Alarm/internal/pkg/Cache"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"

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
	Rcp        *Cache.RedisClientPool
}

func NewListener(url string, kind string, rcp *Cache.RedisClientPool) (*Listener, error) {
	var L Listener
	var err error
	L.Connection, err = messagequeue.NewConnection(url)
	if err != nil {
		return &Listener{}, err
	}
	L.Queue, err = L.Connection.MessageQueueDeclare("hello", false, false, false, false, nil)
	if err != nil {
		return &Listener{}, err
	}

	L.Ans = 0
	L.Control = make(chan bool)
	L.Messages = make(chan []byte)
	L.kind = kind
	L.Rcp = rcp
	L.id = rand.Int()
	return &L, err
}
func (L Listener) Close() error {
	L.Control <- true
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
func (L Listener) Listening(f func([]byte)) (err error) {

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

	switch L.kind {
	case "respone":
		fmt.Println(L.id)
		log.Println(string(body))
		var res Response
		json.Unmarshal(body, &res)
		ts := strings.Split(res.Target, ",")
		for _, t := range ts {
			res.Target = t
			j, err := json.Marshal(res)
			if err != nil {
				fmt.Println(err)
			}
			conn := L.Rcp.GetConn()
			conn.Do("RPUSH", res.Target, string(j))
			conn.Close()
		}
	case "runresult":
		conn := L.Rcp.GetConn()
		conn.Do("RPUSH", "RunResult", string(body))
		conn.Close()
	}

}
