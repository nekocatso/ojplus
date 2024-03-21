package controllers

import "Alarm/internal/pkg/listener"

type ListenerPool struct {
	MaxListener  int
	NumListener  int
	ListenerList []*listener.Listener
}

func NewListenerPool() (*ListenerPool, error) {
	lp := &ListenerPool{MaxListener: 2, NumListener: 0, ListenerList: []*listener.Listener{}}
	for i := 0; i < lp.MaxListener; i++ {
		li, err := listener.NewListener("amqp://user:mkjsix7@172.16.0.15:5672/")
		if err != nil {
			return &ListenerPool{}, err
		}
		lp.ListenerList = append(lp.ListenerList, li)
	}
	return lp, nil
}
func (p *ListenerPool) Run() {
	for _, i := range p.ListenerList {
		i.Listening()
	}
}
