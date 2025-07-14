package middleware

import (
	"context"
	"net/http"
	"strconv"
)

type contextKey string

const TelegramIDKey = contextKey("telegram_id")

func TelegramIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		telegramIDStr := r.Header.Get("X-Telegram-ID")
		if telegramIDStr == "" {
			http.Error(w, "Missing X-Telegram-ID header", http.StatusUnauthorized)
			return
		}

		telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid X-Telegram-ID", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), TelegramIDKey, telegramID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
