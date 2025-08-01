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
	"url-shorter-bot/pkg/logger"
	"url-shorter-bot/pkg/middleware"
	"url-shorter-bot/pkg/migration"
	"url-shorter-bot/pkg/models"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/acme/autocert"
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
		ok, err := migrator.TableExists(table)
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

	//important variablse
	cache := cache.NewMemoryCache(10*time.Minute, 20*time.Minute)
	database := database.NewClient(databaseUrl, databaseApiKey)
	logger := logger.NewDatabaseLogger(database)

	//start bot

	state := bot.NewStateStore()

	handler, err := bot.NewBotHandler(botToken, state, database, logger)
	if err != nil {
		log.Fatalf("❌ Failed to create bot: %v", err)
	}

	go handler.Run()

	//start server
	go middleware.CleanupVisitors()

	port := models.Config.Port
	domain := models.Config.HostName

	shorterUrlHandler := handlers.NewShortdUrlHandler(database, logger)
	hashedUrlHandler := handlers.NewHashedUrlHandler(cache, database, logger)

	r := mux.NewRouter()

	go r.Handle("/short", middleware.TelegramIDMiddleware(http.HandlerFunc(shorterUrlHandler.HandlerUrlShort)))
	go r.HandleFunc("/{url:[0-9]+}", hashedUrlHandler.HandlerHashUrl)

	r.Use(middleware.RateLimitMiddleware)

	switch port {
	case "443":
		if domain == "" {
			log.Fatal("❌ Domain must be set in config.yaml for HTTPS via autocert")
		}

		certManager := &autocert.Manager{
			Cache:      autocert.DirCache("certs"),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(domain),
		}

		go func() {
			err := http.ListenAndServe(":80", certManager.HTTPHandler(nil))
			if err != nil {
				log.Fatalf("❌ HTTP challenge server failed: %v", err)
			}
		}()

		server := &http.Server{
			Addr:      ":443",
			Handler:   r,
			TLSConfig: certManager.TLSConfig(),
		}

		if err := server.ListenAndServeTLS("", ""); err != nil {
			log.Fatalf("❌ HTTPS server failed: %v", err)
		} else {
			fmt.Println("Server is listening")
		}

	default:
		if err := http.ListenAndServe(":"+models.Config.Port, r); err != nil {
			log.Fatalf("❌ HTTP challenge server failed: %v", err)
		} else {
			fmt.Println("Server is listening")
		}
	}
}
