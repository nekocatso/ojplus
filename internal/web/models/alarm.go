package models

import "time"

type AlarmTemplate struct {
	ID        int       `xorm:"'id' pk autoincr"`
	Name      string    `xorm:"notnull unique(name_creator)"`
	Interval  int       `xorm:"notnull"`
	Mails     []string  `xorm:"notnull"`
	CreatorID int       `xorm:"'creator_id' notnull unique(name_creator)"`
	Note      *string   `xorm:"'note'"`
	CreatedAt time.Time `xorm:"'created_at' created"`
	UpdatedAt time.Time `xorm:"'updated_at' updated"`
	DeletedAt time.Time `xorm:"deleted"`
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
