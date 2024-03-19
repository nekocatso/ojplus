package models

import "time"

type User struct {
	ID        int       `xorm:"'id' pk autoincr"`
	Name      string    `xorm:"'name' notnull"`
	Email     string    `xorm:"'email' notnull unique"`
	CreatedAt time.Time `xorm:"'created_at' created"`
	UpdatedAt time.Time `xorm:"'updated_at' updated"`
}
