package main

import (
	"net/http"
	"time"
	"url-shorter-bot/pkg/app/handlers"
	"url-shorter-bot/pkg/cache"
	"url-shorter-bot/pkg/database"
	"url-shorter-bot/pkg/models"

	"github.com/gorilla/mux"
)

func main() {
	models.ReadConfig()

	cache := cache.NewMemoryCache(10*time.Minute, 20*time.Minute)
	database := database.NewClient(models.Config.DatabasebUrl, models.Config.DatabaseApiKey)

	shorterUrlHandler := handlers.NewShortdUrlHandler(database)
	hashedUrlHandler := handlers.NewHashedUrlHandler(cache, database)

	r := mux.NewRouter()

	r.HandleFunc("/short", shorterUrlHandler.HandlerUrlShort)
	r.HandleFunc("/{url:[0-9]+}", hashedUrlHandler.HandlerHashUrl)

	http.ListenAndServe(":"+models.Config.Port, r)
}
