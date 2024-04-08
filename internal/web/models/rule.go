package models

type Rule struct {
	ID          int    `xorm:"'id' pk autoincr"`
	Name        string `xorm:"notnull"`
	CreatorID   int    `xorm:"notnull"`
	AlarmID     int    `xorm:"notnull"`
	Overtime    int    `xorm:"notnull"`
	Interval    int    `xorm:"notnull"`
	WrongLimit  int    `xorm:"notnull"`
	HealthLimit int    `xorm:"notnull"`
}

type PingInfo struct {
	ID           int `xorm:"'id' pk autoincr"`
	Mode         int `xorm:"notnull"`
	LatencyLimit int `xorm:"notnull"`
	LostLimit    int `xorm:"notnull"`
}

type TcpInfo struct {
	ID           int    `xorm:"'id' pk autoincr"`
	EnablePorts  string `xorm:"notnull"`
	DisablePorts string `xorm:"notnull"`
}
