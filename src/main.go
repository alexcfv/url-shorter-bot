package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"url-shorter-bot/pkg/app/bot"
	"url-shorter-bot/pkg/app/handlers"
	"url-shorter-bot/pkg/cache"
	"url-shorter-bot/pkg/database"
	"url-shorter-bot/pkg/middleware"
	"url-shorter-bot/pkg/migration"
	"url-shorter-bot/pkg/models"

	"github.com/gorilla/mux"
)

func main() {
	//read yaml config
	models.ReadConfig()

	//data from config
	databaseUrl := models.Config.DatabasebUrl
	databaseApiKey := models.Config.DatabaseApiKey

	if databaseApiKey == "" || databaseUrl == "" {
		log.Fatal("❌ db_url or db_key is not set")
	}

	botToken := models.Config.TelegramApiKey

	if botToken == "" {
		log.Fatal("❌ tg_key is not set")
	}

	//migrations
	migrator := migration.NewMigrator(databaseUrl, models.Config.DatabaseApiKey)

	for table, request := range models.SqlRequests {
		ok, err := migrator.TablesExists(table)
		if err != nil {
			log.Fatalf("failed to check table: %v", err)
		}

		if !ok {
			fmt.Println("Table " + table + " does not exist. Creating...")
			if err := migrator.CreateTable(table, request); err != nil {
				log.Fatalf("failed to create tables: %v", err)
			}
			fmt.Println("Table " + table + " created")
		} else {
			fmt.Println("Table " + table + " already exists")
		}
	}

	//start bot

	state := bot.NewStateStore()

	handler, err := bot.NewBotHandler(botToken, state)
	if err != nil {
		log.Fatalf("❌ Failed to create bot: %v", err)
	}

	go handler.Run()

	//start server

	cache := cache.NewMemoryCache(10*time.Minute, 20*time.Minute)
	database := database.NewClient(databaseUrl, databaseApiKey)

	shorterUrlHandler := handlers.NewShortdUrlHandler(database)
	hashedUrlHandler := handlers.NewHashedUrlHandler(cache, database)

	r := mux.NewRouter()

	r.HandleFunc("/short", shorterUrlHandler.HandlerUrlShort)
	r.HandleFunc("/{url:[0-9]+}", hashedUrlHandler.HandlerHashUrl)

	r.Use(middleware.RateLimitMiddleware)

	http.ListenAndServe(":"+models.Config.Port, r)
}
