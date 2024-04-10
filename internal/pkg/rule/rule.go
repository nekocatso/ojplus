package rule

import (
	"Alarm/internal/pkg/mail"
	"Alarm/internal/web/models"
	"time"
)

type Rule interface {
	Scan() error
	Update()
}
type State struct {
	nor            int
	abn            int
	Status         int
	reason         string
	correlation_id int
	time           time.Time
}
type tools struct {
	Rcp  *models.Cache
	mail *mail.MailPool
	db   *models.Database
}

//var AlarmCodeStatus = map[int]string{1: "异常触发", 2: "异常持续", 3: "异常结束", 4: "告警中止"}
