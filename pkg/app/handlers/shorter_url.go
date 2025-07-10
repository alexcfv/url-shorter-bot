package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"url-shorter-bot/pkg/app/validators"
	"url-shorter-bot/pkg/database"
	"url-shorter-bot/pkg/models"
)

type UrlShortHandler struct {
	db database.SupabaseClient
}

func NewShortdUrlHandler(db database.SupabaseClient) *UrlShortHandler {
	return &UrlShortHandler{db: db}
}

func (h *UrlShortHandler) HandlerUrlShort(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "must be only POST", http.StatusMethodNotAllowed)
		return
	}

	var reqData models.RequestData
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil || !validators.IsValidURL(reqData.Url) {
		http.Error(w, "invalid JSON body", http.StatusUnsupportedMediaType)
		return
	}

	if ok := validators.IsValidURL(reqData.Url); ok {
		hashUrl := validators.ShortToHash(reqData.Url)
		hashUrlString := strconv.Itoa(int(hashUrl))

		_, err := h.db.Insert("urls", models.Url{Hash: hashUrlString, Url: reqData.Url})
		if err != nil {
			http.Error(w, err.Error(), http.StatusMethodNotAllowed)
			return
		}
		fmt.Fprintf(w, "http://%s:%s/%s", models.Config.HostName, models.Config.Port, hashUrlString)
	} else {
		http.Error(w, "invalid URL", http.StatusUnsupportedMediaType)
	}
}
