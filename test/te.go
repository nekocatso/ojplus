package main

import (
	"Alarm/internal/pkg/listener"
	"Alarm/internal/pkg/messagequeue"
	"fmt"
	"runtime"
	"time"
)

func f(n int, ch chan bool) bool {
	for {
		select {
		case <-ch:
			fmt.Println(n)
			return false
		}
	}

}
func main() {
	ch := make(chan bool)
	for i := 0; i < 100; i++ {
		go f(i, ch)
	}
	time.Sleep(2 * time.Second)
	close(ch)
	time.Sleep(10 * time.Second)
}
func main1() {
	a := runtime.NumCPU()
	b := runtime.NumGoroutine()
	fmt.Print(a, b)
}

func main2() {
	L, _ := listener.NewListener("amqp://user:mkjsix7@172.16.0.15:5672/")
	defer L.Logout()
	L.Listening()
	c, _ := messagequeue.NewConnection("amqp://user:mkjsix7@172.16.0.15:5672/")
	q, _ := c.MessageQueueDeclare("hello", // 队列名称
		false, // 是否持久化
		false, // 是否自动删除
		false, // 是否独占
		false, // 是否等待服务器的响应
		nil,   // 其他参数
	)
	for i := 1; i < 100; i++ {
		q.SendMessage([]byte(fmt.Sprintf("number%d\n", i)))
	}
	time.Sleep(10 * time.Second)
	L.Stop()

}
