package handlers

import (
	"fmt"
	"net/http"
	"time"

	"url-shorter-bot/pkg/cache"

	"github.com/gorilla/mux"
)

type UrlHandler struct {
	cache cache.Cache
}

func NewHashedUrlHandler(c cache.Cache) *UrlHandler {
	return &UrlHandler{cache: c}
}

func (h *UrlHandler) HandlerHashUrl(w http.ResponseWriter, r *http.Request) {
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

	h.cache.Set(hashUrl, "originalUrl", 10*time.Minute)

	fmt.Fprintf(w, "Fetched and cached: %s", hashUrl)
}
