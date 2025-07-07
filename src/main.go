package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	http.ListenAndServe("8000", r)
}
