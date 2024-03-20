package controllers

import "Alarm/internal/services"

type ListenerPool struct {
	MaxListener  int
	NumListener  int
	ListenerList []*services.Listener
}

func NewListenerPool() (*ListenerPool, error) {
	lp := &ListenerPool{MaxListener: 2, NumListener: 0, ListenerList: []*services.Listener{}}
	for i := 0; i < lp.MaxListener; i++ {
		li, err := services.NewListener("amqp://user:mkjsix7@172.16.0.15:5672/")
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
