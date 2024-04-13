package models

import "time"

type AlarmTemplate struct {
	ID        int       `json:"id" xorm:"'id' pk autoincr"`
	Name      string    `json:"name" xorm:"notnull unique(name_creator)"`
	Interval  int       `json:"interval" xorm:"notnull"`
	Mails     []string  `json:"mails" xorm:"notnull"`
	CreatorID int       `json:"creatorID" xorm:"'creator_id' notnull unique(name_creator)"`
	Note      *string   `json:"note" xorm:"'note'"`
	CreatedAt time.Time `json:"createdAt" xorm:"'created_at' created"`
	UpdatedAt time.Time `json:"-" xorm:"'updated_at' updated"`
	DeletedAt time.Time `json:"-" xorm:"deleted"`
}

type AlarmLog struct {
	ID        int       `xorm:"'id' pk autoincr"`
	AssetID   int       `xorm:"'asset_id' notnull"`
	RuleID    int       `xorm:"'rule_id' notnull"`
	State     int       `xorm:"notnull"`
	Mails     []Mail    `xorm:"notnull"`
	Messages  []string  `xorm:"notnull"`
	CreatedAt time.Time `xorm:"'created_at' created"`
}

type Mail struct {
	Address string
	State   bool
}
