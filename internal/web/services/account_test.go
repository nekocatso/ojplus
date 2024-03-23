package services

import (
	"Alarm/internal/config"
	"Alarm/internal/web/models"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

var (
	prepareTestData = func() *models.User {
		return &models.User{
			Username: "testUser",
			Name:     "测试员",
			Password: "testPassword",
			Email:    "test@example.com",
			Phone:    "1234567890",
		}
	}
	mysqlConfig = &config.MysqlConfig{
		DSN: "zzh:123.com@tcp(127.0.0.1:3306)/alarm?charset=utf8",
	}
	db, _ = models.NewDatabase(mysqlConfig)
	svc   = &Account{db: db}
	user  = prepareTestData()
)

func TestAccountService_DeleteUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		oneUser := &models.User{ID: 1}
		err := svc.DeleteUser(oneUser)
		if err != nil {
			t.Errorf("AccountService.CreateUser() error = %v, wantErr %v", err, false)
		}
	})

}
func TestAccountService_CreateUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		err := svc.CreateUser(user)
		if err != nil {
			t.Errorf("AccountService.CreateUser() error = %v, wantErr %v", err, false)
		}
	})

}
