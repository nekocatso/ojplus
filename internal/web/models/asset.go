package models

import "time"

type Asset struct {
	ID        int       `xorm:"'id' pk autoincr"`
	Name      string    `xorm:"notnull unique(name_creator)"`
	Type      string    `xorm:"notnull"`
	Address   string    `xorm:"'address' notnull unique"`
	Note      string    `xorm:"'note'"`
	State     int       `xorm:"default(0) notnull"`
	CreatorID int       `xorm:"'creator_id' notnull unique(name_creator)"`
	CreatedAt time.Time `xorm:"'created_at' created"`
	UpdatedAt time.Time `xorm:"'updated_at' updated"`
	DeletedAt time.Time `xorm:"deleted"`
	Version   int       `xorm:"version"`
}

type AssetUser struct {
	ID      int `xorm:"'id' pk autoincr"`
	AssetID int `xorm:"'asset_id' notnull unique(asset_user)"`
	UserID  int `xorm:"'user_id' notnull unique(asset_user)"`
}

type AssetRule struct {
	ID      int `xorm:"'id' pk autoincr"`
	AssetID int `xorm:"'asset_id' notnull unique(asset_rule)"`
	RuleID  int `xorm:"'rule_id' notnull unique(asset_rule)"`
}
