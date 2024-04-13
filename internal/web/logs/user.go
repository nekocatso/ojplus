package logs

import (
	"Alarm/internal/web/models"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

func (l *Logger) SaveUserLog(ctx *gin.Context, userLog *UserLog) error {
	userID := getUserIDByContext(ctx)
	if userID == 0 {
		return errors.New("user id cannot be 0")
	}
	_, err := l.db.Engine.Insert(&models.UserLog{
		UserID:  userID,
		Module:  userLog.Module,
		Type:    userLog.Type,
		Content: userLog.Content,
		IP:      ctx.ClientIP(),
	})
	return err
}

func getUserIDByContext(ctx *gin.Context) int {
	claims := ctx.Value("claims").(jwt.MapClaims)
	return claims["userID"].(int)
}
