package models

import "time"

type User struct {
	ID        int       `xorm:"'id' pk autoincr"`
	Username  string    `xorm:"notnull"`
	Name      string    `xorm:"notnull"`
	Password  string    `xorm:"notnull"`
	Email     string    `xorm:"null"`
	Phone     string    `xorm:"null"`
	CreatedAt time.Time `xorm:"created"`
	UpdatedAt time.Time `xorm:"updated"`
	DeletedAt time.Time `xorm:"deleted"`
	Version   int       `xorm:"version"`
}

func (u *User) GetInfo() *UserInfo {
	return &UserInfo{
		ID:        u.ID,
		Username:  u.Username,
		Name:      u.Name,
		Email:     u.Email,
		Telephone: u.Phone,
	}
}

type UserInfo struct {
	ID        int
	Username  string
	Name      string
	Email     string
	Telephone string
}
