package logger

import (
	"errors"
	"testing"
	"url-shorter-bot/pkg/models"
)

type mockInserter struct {
	calledTable string
	calledData  interface{}
	returnErr   error
}

func (m *mockInserter) Insert(table string, data interface{}) ([]byte, error) {
	m.calledTable = table
	m.calledData = data
	return nil, m.returnErr
}

func TestLogAction(t *testing.T) {
	tests := []struct {
		name       string
		telegramID int64
		action     string
		returnErr  error
		wantTable  string
	}{
		{
			name:       "success log action",
			telegramID: 123456,
			action:     "User clicked link",
			returnErr:  nil,
			wantTable:  "log_action",
		},
		{
			name:       "fail to insert action",
			telegramID: 999,
			action:     "Something failed",
			returnErr:  errors.New("db error"),
			wantTable:  "log_action",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockInserter{returnErr: tt.returnErr}
			log := NewDatabaseLogger(mock)

			log.LogAction(tt.telegramID, tt.action)

			if mock.calledTable != tt.wantTable {
				t.Errorf("expected table %s, got %s", tt.wantTable, mock.calledTable)
			}

			gotData, ok := mock.calledData.(models.LogAction)
			if !ok {
				t.Errorf("unexpected payload type: %T", mock.calledData)
			}
			if gotData.Telegram_id != tt.telegramID || gotData.Action != tt.action {
				t.Errorf("unexpected payload data: %+v", gotData)
			}
		})
	}
}

func TestLogError(t *testing.T) {
	tests := []struct {
		name       string
		telegramID int64
		errMsg     string
		errCode    string
		returnErr  error
		wantTable  string
	}{
		{
			name:       "success log error",
			telegramID: 321,
			errMsg:     "bad request",
			errCode:    "400",
			returnErr:  nil,
			wantTable:  "log_error",
		},
		{
			name:       "fail to log error",
			telegramID: 321,
			errMsg:     "timeout",
			errCode:    "504",
			returnErr:  errors.New("network fail"),
			wantTable:  "log_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockInserter{returnErr: tt.returnErr}
			log := NewDatabaseLogger(mock)

			log.LogError(tt.telegramID, tt.errMsg, tt.errCode)

			if mock.calledTable != tt.wantTable {
				t.Errorf("expected table %s, got %s", tt.wantTable, mock.calledTable)
			}

			gotData, ok := mock.calledData.(models.LogError)
			if !ok {
				t.Errorf("unexpected payload type: %T", mock.calledData)
			}
			if gotData.Telegram_id != tt.telegramID || gotData.Error != tt.errMsg || gotData.Error_code != tt.errCode {
				t.Errorf("unexpected payload data: %+v", gotData)
			}
		})
	}
}
