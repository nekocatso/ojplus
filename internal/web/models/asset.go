package models

import "time"

type Asset struct {
	ID        int       `json:"id" xorm:"'id' pk autoincr"`
	Name      string    `json:"name" xorm:"notnull unique(name_creator)"`
	Type      string    `json:"type" xorm:"notnull"`
	Address   string    `json:"address" xorm:"'address' notnull unique"`
	Note      string    `json:"note" xorm:"'note'"`
	State     int       `json:"state" xorm:"default(0) notnull"`
	CreatorID int       `json:"-" xorm:"'creator_id' notnull unique(name_creator)"`
	CreatedAt time.Time `json:"createdAt" xorm:"'created_at' created"`
	UpdatedAt time.Time `json:"-" xorm:"'updated_at' updated"`
	DeletedAt time.Time `json:"-" xorm:"deleted"`
	Creator   *UserInfo `json:"creator" xorm:"-"`
}

type AssetUser struct {
	ID        int       `xorm:"'id' pk autoincr"`
	AssetID   int       `xorm:"'asset_id' notnull unique(asset_user)"`
	UserID    int       `xorm:"'user_id' notnull unique(asset_user)"`
	CreatedAt time.Time `json:"-" xorm:"'created_at' created"`
}

type AssetRule struct {
	ID        int       `xorm:"'id' pk autoincr"`
	AssetID   int       `xorm:"'asset_id' notnull unique(asset_rule)"`
	RuleID    int       `xorm:"'rule_id' notnull unique(asset_rule)"`
	CreatedAt time.Time `json:"-" xorm:"'created_at' created"`
}
