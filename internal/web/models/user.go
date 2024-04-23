package models

import (
	"fmt"
	"log"
	"reflect"
	"time"
)

type User struct {
	ID        int       `json:"id" xorm:"'id' pk autoincr"`
	Username  *string   `json:"username" xorm:" notnull unique(username)"`
	Nickname  *string   `json:"nickname" xorm:"null"`
	Name      *string   `json:"name" xorm:"null"`
	Password  *string   `json:"-" xorm:"notnull"`
	Email     *string   `json:"email" xorm:"notnull unique(email)"`
	Role      *int      `json:"role" xorm:"default(10)"`
	Disable   *bool     `json:"disable" xorm:"default(false)"`
	CreatedAt time.Time `json:"createdAt" xorm:"created"`
	UpdatedAt time.Time `json:"-" xorm:"updated"`
	DeletedAt time.Time `json:"-" xorm:"deleted unique(username) unique(email)"`
}

type UserManager struct {
	db *Database
}

func NewUserManager(db *Database) *UserManager {
	return &UserManager{db: db}
}

func (manager *UserManager) Create(model any) (int, error) {
	_, err := manager.db.Engine.Table(new(User)).Insert(model)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	// 获取userID
	value := reflect.ValueOf(model).Elem()
	field := value.FieldByName("Username")
	if !field.IsValid() {
		return 0, fmt.Errorf("no Username field found in model")
	}
	username := field.Interface().(*string)
	log.Println(*username)
	var userID int
	_, err = manager.db.Engine.Table(new(User)).Where("username = ?", *username).Cols("id").Get(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (manager *UserManager) Update(userID int, model any) error {
	_, err := manager.db.Engine.Table(new(User)).Where("id = ?", userID).Update(model)
	return err
}

func (manager *UserManager) Delete(userID int) error {
	user, err := manager.GetByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}
	_, err = manager.db.Engine.ID(userID).Delete(&User{})
	return err
}

func (manager *UserManager) GetByID(userID int) (*User, error) {
	if userID <= 0 {
		return nil, nil
	}
	user := new(User)
	_, err := manager.db.Engine.ID(userID).Get(user)
	return user, err
}

func (manager *UserManager) GetByPage(page, pageSize int) ([]User, error) {
	var users []User
	offset := (page - 1) * pageSize
	err := manager.db.Engine.Limit(pageSize, offset).Find(&users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (manager *UserManager) GetBySelf(user *User) (bool, error) {
	return manager.db.Engine.Get(user)
}

func (manager *UserManager) Count(query interface{}, args ...interface{}) (int64, error) {
	return manager.db.Engine.Where(query, args...).Count(&User{})
}
