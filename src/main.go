package main

import (
	"net/http"
	"time"
	"url-shorter-bot/pkg/app/handlers"
	"url-shorter-bot/pkg/cache"
	"url-shorter-bot/pkg/models"

	"github.com/gorilla/mux"
)

func main() {
	models.ReadConfig()

	c := cache.NewMemoryCache(10*time.Minute, 20*time.Minute)
	hashedUrlHandler := handlers.NewHashedUrlHandler(c)

	r := mux.NewRouter()

	r.HandleFunc("/short", handlers.HandlerUrlShort)
	r.HandleFunc("/{url:[0-9]+}", hashedUrlHandler.HandlerHashUrl)

	http.ListenAndServe(":"+models.Config.Port, r)
}
