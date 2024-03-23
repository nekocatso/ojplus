package models

import "time"

type Token struct {
	ID        int       `xorm:"'id' pk autoincr"`
	UserID    int       `xorm:"'user_id' notnull unique"`
	Content   string    `xorm:"'content' notnull"`
	ExpiredAt time.Time `xorm:"'expired_at'`
	CreatedAt time.Time `xorm:"'created_at' created"`
	UpdatedAt time.Time `xorm:"'updated_at' updated"`
}
