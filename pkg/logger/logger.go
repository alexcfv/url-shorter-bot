package logger

type Logger interface {
	LogAction(telegramID int64, action string)
	LogError(telegramID int64, errMsg, code string)
}
