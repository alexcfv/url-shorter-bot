package main

import (
	"net/http"
	"url-shorter-bot/pkg/app/handlers"
	"url-shorter-bot/pkg/models"

	"github.com/gorilla/mux"
)

func main() {
	models.Config = models.ReadConfig()

	r := mux.NewRouter()

	r.HandleFunc("/short", handlers.HandlerUrlShort)

	http.ListenAndServe(":"+models.Config.Port, r)
}
