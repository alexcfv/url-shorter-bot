package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"url-shorter-bot/pkg/cache"
	"url-shorter-bot/pkg/database"

	"github.com/gorilla/mux"
)

type UrlHashHandler struct {
	cache cache.Cache
	db    database.SupabaseClient
}

func NewHashedUrlHandler(c cache.Cache, db database.SupabaseClient) *UrlHashHandler {
	return &UrlHashHandler{cache: c, db: db}
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

	if val, found := h.cache.Get(hashUrl); found {
		fmt.Fprintf(w, "From cache: %s", val)
		return
	}

	valBytes, err := h.db.Get("urls", map[string]string{
		"hash": hashUrl,
	})
	if err != nil {
		http.Error(w, "Not found url", http.StatusBadRequest)
		return
	}

	var result struct {
		Hash        string `json:"hash"`
		OriginalUrl string `json:"original_url"`
	}
	if err := json.Unmarshal(valBytes, &result); err != nil {
		http.Error(w, "invalid data from DB", http.StatusInternalServerError)
		return
	}

	h.cache.Set(hashUrl, result.OriginalUrl, 10*time.Minute)
	fmt.Fprintf(w, "Fetched and cached: %s", result.OriginalUrl)
}
