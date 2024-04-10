package models

import "time"

type Rule struct {
	ID           int    `xorm:"'id' pk autoincr"`
	Name         string `xorm:"notnull  unique(name_creator)"`
	Type         string `xorm:"notnull"`
	CreatorID    int    `xorm:"'creator_id' notnull unique(name_creator)"`
	AlarmID      int    `xorm:"'alarm_id' notnull"`
	Overtime     int    `xorm:"notnull"`
	Interval     int    `xorm:"notnull"`
	DeclineLimit int    `xorm:"notnull"`
	RecoverLimit int    `xorm:"notnull"`
	Note         string
	CreateAt     time.Time   `json:"-" xorm:"'created_at' created"`
	UpdateAt     time.Time   `json:"-" xorm:"'updated_at' updated"`
	DeleteAt     time.Time   `json:"-" xorm:"deleted"`
	Creator      *UserInfo   `json:"creator" xorm:"-"`
	Info         interface{} `json:"info" xorm:"-"`
}

type PingInfo struct {
	ID           int       `xorm:"'id' pk"`
	Mode         int       `xorm:"notnull"`
	LatencyLimit int       `xorm:"notnull"`
	LostLimit    int       `xorm:"notnull"`
	CreateAt     time.Time `json:"-" xorm:"'created_at' created"`
	UpdateAt     time.Time `json:"-" xorm:"'updated_at' updated"`
	DeleteAt     time.Time `json:"-" xorm:"deleted"`
}

type TCPInfo struct {
	ID           int       `xorm:"'id' pk"`
	EnablePorts  string    `xorm:"notnull"`
	DisablePorts string    `xorm:"notnull"`
	CreateAt     time.Time `json:"-" xorm:"'created_at' created"`
	UpdateAt     time.Time `json:"-" xorm:"'updated_at' updated"`
	DeleteAt     time.Time `json:"-" xorm:"deleted"`
}

func (t *TCPInfo) TableName() string {
	return "tcp_info"
}
