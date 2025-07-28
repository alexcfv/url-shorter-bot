package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"url-shorter-bot/pkg/cache"
	"url-shorter-bot/pkg/database"
	"url-shorter-bot/pkg/logger"
	"url-shorter-bot/pkg/models"

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

	hashUrl := mux.Vars(r)["url"]
	if hashUrl == "" {
		http.Error(w, "missing hash url", http.StatusBadRequest)
		return
	}

	if cachedUrl, ok := h.cache.Get(hashUrl); ok {
		cachedUrlString, ok := cachedUrl.(string)
		if !ok {
			http.Error(w, "Error", http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, cachedUrlString, http.StatusFound)
		return
	}

	valBytes, err := h.db.Get("urls", map[string]string{
		"Hash": hashUrl,
	})
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var result models.Url

	if err := json.Unmarshal(valBytes, &result); err != nil {
		http.Error(w, "invalid data from DB", http.StatusInternalServerError)
		return
	}

	go h.cache.Set(hashUrl, result.Url, 10*time.Minute)

	go func(h *UrlHashHandler) {
		h.logger.LogAction(result.Telegram_id, "users url has been used")
	}(h)

	http.Redirect(w, r, result.Url, http.StatusFound)
}
