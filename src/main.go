package main

import (
	"net/http"
	"url-shorter-bot/pkg/app/handlers"
	"url-shorter-bot/pkg/models"

	"github.com/gorilla/mux"
)

func main() {
	models.ReadConfig()

	r := mux.NewRouter()

	r.HandleFunc("/short", handlers.HandlerUrlShort)
	r.HandleFunc("/{url:[0-9]+}", handlers.HandlerHashUrl)

	http.ListenAndServe(":"+models.Config.Port, r)
}
