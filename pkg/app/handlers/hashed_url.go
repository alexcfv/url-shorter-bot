package handlers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func HandlerHashUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "must be only GET", http.StatusMethodNotAllowed)
		return
	}
	hashUrl := mux.Vars(r)["url"]
	fmt.Fprintf(w, "%s", hashUrl)
}
