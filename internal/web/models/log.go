package models

import "time"

type AlarmLog struct {
	ID        int       `json:"id" xorm:"'id' pk autoincr"`
	AssetID   int       `json:"assetID" xorm:"'asset_id' notnull"`
	RuleID    int       `json:"ruleID" xorm:"'rule_id' notnull"`
	State     int       `json:"state" xorm:"notnull"`
	Mails     []Mail    `json:"mails" xorm:"notnull"`
	Messages  []string  `json:"messages" xorm:"notnull"`
	CreatedAt time.Time `json:"createdAt" xorm:"'created_at' created"`
	AssetName string    `json:"assetName" xorm:"-"`
	RuleName  string    `json:"ruleName" xorm:"-"`
	Admin     string    `json:"admin" xorm:"-"`
	RuleType  string    `json:"ruleType" xorm:"-"`
}

type UserLog struct {
	ID       int       `json:"id" xorm:"'id' pk autoincr"`
	UserID   int       `json:"-" xorm:"'user_id' notnull"`
	Username string    `json:"username" xorm:"notnull"`
	Phone    string    `json:"phone" xorm:"notnull"`
	Module   string    `json:"module" xorm:"notnull"`
	Type     string    `json:"type" xorm:"notnull"`
	Content  string    `json:"content" xorm:"notnull"`
	IP       string    `json:"ip" xorm:"'ip' notnull"`
	CreateAt time.Time `json:"createdAt" xorm:"'created_at' created"`
}
