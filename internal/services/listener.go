package services

import (
	"Alarm/internal/models"
	"errors"
	"fmt"

	"github.com/streadway/amqp"
)

type Listener struct {
	Connection *models.Connection
	Queue      *models.MessageQueue
	Ans        int
	Control    chan bool
	Messages   chan []byte
}

func NewListener(url string) (*Listener, error) {
	var L Listener
	var err error
	L.Connection, err = models.NewConnection(url)

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
	return &L, err
}
func (L Listener) Logout() (err error) {
	if L.Queue == nil {
		return errors.New("L.Queue is nil")
	}
	err = L.Queue.Close()
	if err != nil {
		return
	}
	if L.Connection == nil {
		return errors.New("L.Connection is nil")
	}
	err = L.Connection.Close()
	if err != nil {
		return
	}

	return
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
				Save(msg)
			case <-c:
				return
			}
		}
	}(L.Control, messages)
	return
}
func Save(msg amqp.Delivery) {
	fmt.Printf("%s", msg.Body)
}
