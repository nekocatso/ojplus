package main

import (
	"Alarm/internal/models"
	"Alarm/internal/services"
	"fmt"
)

func main() {
	L, _ := services.NewListener("amqp://user:mkjsix7@172.16.0.15:5672/")
	L.Listening()
	c, _ := models.NewConnection("amqp://user:mkjsix7@172.16.0.15:5672/")
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

}
