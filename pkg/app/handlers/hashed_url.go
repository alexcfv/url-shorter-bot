package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"url-shorter-bot/pkg/cache"
	"url-shorter-bot/pkg/database"
	"url-shorter-bot/pkg/logger"
	"url-shorter-bot/pkg/middleware"

	"github.com/gorilla/mux"
)

type UrlHashHandler struct {
	cache  cache.Cache
	db     database.SupabaseClient
	logger logger.Logger
}

func NewHashedUrlHandler(c cache.Cache, db database.SupabaseClient, log logger.Logger) *UrlHashHandler {
	return &UrlHashHandler{cache: c, db: db, logger: log}
}

func (h *UrlHashHandler) HandlerHashUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "must be only GET", http.StatusMethodNotAllowed)
		return
	}

	telegramIDValue := r.Context().Value(middleware.TelegramIDKey)
	if telegramIDValue == nil {
		http.Error(w, "No telegram_id in context", http.StatusInternalServerError)
		return
	}
	telegramID := telegramIDValue.(int64)

	h.logger.LogAction(telegramID, "go to shorten link")

	hashUrl := mux.Vars(r)["url"]
	if hashUrl == "" {
		h.logger.LogError(telegramID, "missing hash url", "400")
		http.Error(w, "missing hash url", http.StatusBadRequest)
		return
	}

	if val, found := h.cache.Get(hashUrl); found {
		fmt.Fprintf(w, "From cache: %s", val)
		return
	}

	valBytes, err := h.db.Get("urls", map[string]string{
		"Hash": hashUrl,
	})
	if err != nil {
		h.logger.LogError(telegramID, err.Error(), "400")
		http.Error(w, "Not found url", http.StatusBadRequest)
		return
	}

	var result struct {
		Hash        string `json:"hash"`
		OriginalUrl string `json:"original_url"`
	}
	if err := json.Unmarshal(valBytes, &result); err != nil {
		h.logger.LogError(telegramID, err.Error(), "500")
		http.Error(w, "invalid data from DB", http.StatusInternalServerError)
		return
	}

	h.cache.Set(hashUrl, result.OriginalUrl, 10*time.Minute)
	fmt.Fprintf(w, "Fetched and cached: %s", result.OriginalUrl)
}
