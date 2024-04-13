package models

import "time"

type User struct {
	ID        int       `json:"id" xorm:"'id' pk autoincr"`
	Username  string    `json:"usernameusername" xorm:"notnull"`
	Name      string    `json:"name" xorm:"notnull"`
	Password  string    `json:"-" xorm:"notnull"`
	Email     string    `json:"email" xorm:"null"`
	Phone     string    `json:"phone" xorm:"null"`
	Role      int       `json:"role" xorm:"default(10)"`
	IsActive  bool      `json:"isActive" xorm:"default(true)"`
	CreatedAt time.Time `json:"createdAt" xorm:"created"`
	UpdatedAt time.Time `json:"-" xorm:"updated"`
	DeletedAt time.Time `json:"-" xorm:"deleted"`
	Note      *string   `json:"note" xorm:"null"`
	IP        string    `json:"-" xorm:"-"`
}

func (u *User) GetInfo() *UserInfo {
	return &UserInfo{
		ID:        u.ID,
		Username:  u.Username,
		Name:      u.Name,
		Email:     u.Email,
		Phone:     u.Phone,
		Role:      u.Role,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		Note:      u.Note,
	}
}

type UserInfo struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Role      int       `json:"role"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createAt"`
	Note      *string   `json:"note"`
}
