package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"url-shorter-bot/pkg/app/validators"
	"url-shorter-bot/pkg/database"
	"url-shorter-bot/pkg/logger"
	"url-shorter-bot/pkg/middleware"
	"url-shorter-bot/pkg/models"
)

type UrlShortHandler struct {
	db     database.SupabaseClient
	logger logger.Logger
}

func NewShortdUrlHandler(db database.SupabaseClient, log logger.Logger) *UrlShortHandler {
	return &UrlShortHandler{db: db, logger: log}
}

func (h *UrlShortHandler) HandlerUrlShort(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "must be only POST", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "invalid content type", http.StatusUnsupportedMediaType)
		return
	}

	telegramIDValue := r.Context().Value(middleware.TelegramIDKey)
	if telegramIDValue == nil {
		http.Error(w, "No telegram_id in context", http.StatusInternalServerError)
		return
	}
	telegramID := telegramIDValue.(int64)

	go h.logger.LogAction(telegramID, "shortened link")

	var reqData models.RequestData
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		go h.logger.LogError(telegramID, err.Error(), "415")
		http.Error(w, "invalid JSON body", http.StatusUnsupportedMediaType)
		return
	}

	if ok := validators.IsValidURL(reqData.Url); ok {
		hashUrl := validators.ShortToHash(reqData.Url + strconv.Itoa(int(telegramID)))
		hashUrlString := strconv.Itoa(int(hashUrl))

		go func() {
			_, err := h.db.Insert("urls", models.Url{Hash: hashUrlString, Url: reqData.Url})
			if err != nil {
				go h.logger.LogError(telegramID, err.Error(), "405")
				http.Error(w, err.Error(), http.StatusMethodNotAllowed)
				return
			}
		}()

		response := models.Respons{
			Url: fmt.Sprintf("%s://%s/%s", models.Protocol, models.Config.HostName, hashUrlString),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	} else {
		http.Error(w, "invalid URL", http.StatusUnsupportedMediaType)
	}
}
