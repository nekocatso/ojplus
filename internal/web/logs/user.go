package logs

import (
	"Alarm/internal/web/models"

	"github.com/gin-gonic/gin"
)

type Logger struct {
	db *models.Database
}

func NewLogger(db *models.Database) *Logger {
	return &Logger{
		db: db,
	}
}

type UserLog struct {
	Module  string
	Type    string
	Content string
}

func (l *Logger) SaveUserLog(ctx *gin.Context, user *models.User, userLog *UserLog) error {
	_, err := l.db.Engine.Insert(&models.UserLog{
		UserID:   user.ID,
		Username: user.Username,
		Phone:    user.Phone,
		Module:   userLog.Module,
		Type:     userLog.Type,
		Content:  userLog.Content,
		IP:       ctx.ClientIP(),
	})
	return err
}
