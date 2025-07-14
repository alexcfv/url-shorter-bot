package logger

import (
	"log"
	"url-shorter-bot/pkg/models"
)

type SupabaseInserter interface {
	Insert(table string, data interface{}) ([]byte, error)
}

type SupabaseLogger struct {
	db SupabaseInserter
}

func NewDatabaseLogger(db SupabaseInserter) *SupabaseLogger {
	return &SupabaseLogger{db: db}
}

func (l *SupabaseLogger) LogAction(telegramID int64, action string) {
	payload := models.LogAction{
		Telegram_id: telegramID,
		Action:      action,
	}
	if _, err := l.db.Insert("log_action", payload); err != nil {
		log.Printf("log action failed: %v", err)
	}
}

func (l *SupabaseLogger) LogError(telegramID int64, errMsg, code string) {
	payload := models.LogError{
		Telegram_id: telegramID,
		Error:       errMsg,
		Error_code:  code,
	}
	if _, err := l.db.Insert("log_error", payload); err != nil {
		log.Printf("log error failed: %v", err)
	}
}
