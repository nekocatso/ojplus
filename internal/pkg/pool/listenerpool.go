package pool

import (
	"Alarm/internal/pkg/listener"
	"Alarm/internal/pkg/Cache"
	"fmt"
)

type ListenerPool struct {
	MaxListener     int
	NumListener     int
	ListenerList    []*listener.Listener
	RedisClientPool *Cache.RedisClientPool
	
}

func NewListenerPool() (*ListenerPool, error) {
	lp := &ListenerPool{MaxListener: 2, NumListener: 0, ListenerList: []*listener.Listener{},
		RedisClientPool: Cache.NewRedisClientPool()}
	for i := 0; i < lp.MaxListener; i++ {
		li, err := listener.NewListener("amqp://user:mkjsix7@172.16.0.15:5672/", "respone", lp.RedisClientPool)
		if err != nil {
			return &ListenerPool{}, err
		}
		lp.ListenerList = append(lp.ListenerList, li)
	}
	fmt.Println("newListenerpool ok")
	return lp, nil
}
func (p *ListenerPool) Run() {
	for _, i := range p.ListenerList {
		i.Listening(func([]byte) {})
	}
	fmt.Println("run Listenerpool ok")
}
func (p *ListenerPool) Close() {
	for _, i := range p.ListenerList {
		i.Stop()
		i.Close()
	}
	p.RedisClientPool.Close()
}
