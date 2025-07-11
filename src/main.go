package main

import (
	"log"
	"net/http"
	"time"
	"url-shorter-bot/pkg/app/bot"
	"url-shorter-bot/pkg/app/handlers"
	"url-shorter-bot/pkg/cache"
	"url-shorter-bot/pkg/database"
	"url-shorter-bot/pkg/middleware"
	"url-shorter-bot/pkg/models"

	"github.com/gorilla/mux"
)

func main() {
	//start server
	models.ReadConfig()

	databaseUrl := models.Config.DatabasebUrl
	databaseApiKey := models.Config.DatabaseApiKey

	if databaseApiKey == "" || databaseUrl == "" {
		log.Fatal("❌ db_url or db_key is not set")
	}

	cache := cache.NewMemoryCache(10*time.Minute, 20*time.Minute)
	database := database.NewClient(databaseUrl, databaseApiKey)

	shorterUrlHandler := handlers.NewShortdUrlHandler(database)
	hashedUrlHandler := handlers.NewHashedUrlHandler(cache, database)

	r := mux.NewRouter()

	r.HandleFunc("/short", shorterUrlHandler.HandlerUrlShort)
	r.HandleFunc("/{url:[0-9]+}", hashedUrlHandler.HandlerHashUrl)

	r.Use(middleware.RateLimitMiddleware)

	http.ListenAndServe(":"+models.Config.Port, r)

	//start bot
	botToken := models.Config.TelegramApiKey
	apiURL := ""

	if botToken == "" || apiURL == "" {
		log.Fatal("❌ tg_key or tg_url is not set")
	}

	state := bot.NewStateStore()

	handler, err := bot.NewBotHandler(botToken, apiURL, state)
	if err != nil {
		log.Fatalf("❌ Failed to create bot: %v", err)
	}

	go handler.Run()
}
