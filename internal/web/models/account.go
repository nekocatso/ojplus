package models

import "time"

type User struct {
	ID        int       `json:"id" xorm:"'id' pk autoincr"`
	Username  string    `json:"username" xorm:"notnull unique(username)"`
	Name      string    `json:"name" xorm:"notnull"`
	Password  string    `json:"-" xorm:"notnull"`
	Email     string    `json:"email" xorm:"null"`
	Phone     string    `json:"phone" xorm:"null"`
	Role      int       `json:"role" xorm:"default(10)"`
	IsActive  bool      `json:"isActive" xorm:"default(true)"`
	CreatedAt time.Time `json:"createdAt" xorm:"created"`
	UpdatedAt time.Time `json:"-" xorm:"updated"`
	DeletedAt time.Time `json:"-" xorm:"deleted unique(username)"`
	Note      *string   `json:"note" xorm:"null"`
	IP        string    `json:"-" xorm:"-"`
}
