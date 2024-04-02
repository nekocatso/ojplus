package models

import "time"

type Asset struct {
	ID        int       `xorm:"'id' pk autoincr"`
	Name      string    `xorm:"'name' notnull"`
	Address   string    `xorm:"'address' notnull"`
	Domain    string    `xorm:"'domain'"`
	CreatorID int       `xorm:"'creator_id' notnull"`
	CreatedAt time.Time `xorm:"'created_at' created"`
	UpdatedAt time.Time `xorm:"'updated_at' updated"`
	DeletedAt time.Time `xorm:"deleted"`
	Version   int       `xorm:"version"`
}
