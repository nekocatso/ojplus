package models

import "time"

type User struct {
	ID        int       `xorm:"'id' pk autoincr"`
	Username  string    `xorm:"notnull"`
	Name      string    `xorm:"notnull"`
	Password  string    `xorm:"notnull"`
	Email     string    `xorm:"null"`
	Phone     string    `xorm:"null"`
	Role      int       `xorm:"default(10)"`
	IsActive  bool      `xorm:"default(true)"`
	CreatedAt time.Time `xorm:"created"`
	UpdatedAt time.Time `xorm:"updated"`
	DeletedAt time.Time `xorm:"deleted"`
	Note      *string   `xorm:"null"`
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
}
