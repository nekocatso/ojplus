package models

import "time"

type Rule struct {
	ID           int         `json:"id" xorm:"'id' pk autoincr"`
	Name         string      `json:"name" xorm:"notnull  unique(name_creator)"`
	Type         string      `json:"type" xorm:"notnull"`
	CreatorID    int         `json:"-" xorm:"'creator_id' notnull unique(name_creator)"`
	AlarmID      int         `json:"alarmID" xorm:"'alarm_id' notnull"`
	Overtime     int         `json:"overtime" xorm:"notnull"`
	Interval     int         `json:"interval" xorm:"notnull"`
	DeclineLimit int         `json:"declineLimit" xorm:"notnull"`
	RecoverLimit int         `json:"recoverLimit" xorm:"notnull"`
	Note         *string     `json:"note" xorm:"null"`
	CreateAt     time.Time   `json:"createAt" xorm:"'created_at' created"`
	UpdateAt     time.Time   `json:"-" xorm:"'updated_at' updated"`
	DeleteAt     time.Time   `json:"-" xorm:"deleted"`
	Creator      *User       `json:"creator" xorm:"-"`
	AssetNames   []string    `json:"assetNames" xorm:"-"`
	AssetsCount  int         `json:"assetsCount" xorm:"-"`
	Info         interface{} `json:"info" xorm:"-"`
}

type PingInfo struct {
	ID           int       `json:"-" xorm:"'id' pk"`
	Mode         int       `json:"mode" xorm:"notnull"`
	LatencyLimit int       `json:"latencyLimit" xorm:"notnull"`
	LostLimit    int       `json:"lostLimit" xorm:"notnull"`
	CreateAt     time.Time `json:"-" xorm:"'created_at' created"`
	UpdateAt     time.Time `json:"-" xorm:"'updated_at' updated"`
	DeleteAt     time.Time `json:"-" xorm:"deleted"`
}

type TCPInfo struct {
	ID           int       `json:"-" xorm:"'id' pk"`
	EnablePorts  string    `json:"enablePorts" xorm:"notnull"`
	DisablePorts string    `json:"disablePorts" xorm:"notnull"`
	CreateAt     time.Time `json:"-" xorm:"'created_at' created"`
	UpdateAt     time.Time `json:"-" xorm:"'updated_at' updated"`
	DeleteAt     time.Time `json:"-" xorm:"deleted"`
}

func (t *TCPInfo) TableName() string {
	return "tcp_info"
}
