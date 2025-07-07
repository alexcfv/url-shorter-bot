package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"url-shorter-bot/pkg/app/validators"
	"url-shorter-bot/pkg/models"
)

func HandlerUrlShort(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "must be only POST", http.StatusMethodNotAllowed)
		return
	}

	var reqData models.RequestData
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil || !validators.IsValidURL(reqData.Url) {
		http.Error(w, "invalid JSON body", http.StatusUnsupportedMediaType)
		return
	}

	if ok := validators.IsValidURL(reqData.Url); ok {
		hashUrl := validators.ShortToHash(reqData.Url)
		hashUrlString := strconv.Itoa(int(hashUrl))

		fmt.Fprintf(w, "http://%s:%s/%s", models.Config.HostName, "8000", hashUrlString)
	} else {
		http.Error(w, "invalid URL", http.StatusUnsupportedMediaType)
	}
}
