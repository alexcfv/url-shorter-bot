// internal/bot/handler.go
package bot

import (
	"encoding/json"
	"net/http"
	"strings"
	"url-shorter-bot/pkg/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotHandler struct {
	Bot   *tgbotapi.BotAPI
	State *StateStore
}

func NewBotHandler(token string, state *StateStore) (*BotHandler, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &BotHandler{Bot: bot, State: state}, nil
}

func (h *BotHandler) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := h.Bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			chatID := update.Message.Chat.ID
			text := update.Message.Text

			switch {
			case text == "/start":
				msg := tgbotapi.NewMessage(chatID, "👋 Welcome! Click the button below to shorten a URL.")
				msg.ReplyMarkup = UrlShortenKeyboard()
				h.Bot.Send(msg)

			case text == "Shorten URL":
				h.State.Set(chatID, "awaiting_url")
				msg := tgbotapi.NewMessage(chatID, "Please send the URL you want to shorten.")
				h.Bot.Send(msg)

			case h.State.Get(chatID) == "awaiting_url":
				h.State.Clear(chatID)
				shortURL, err := h.shortenURL(text)
				if err != nil || shortURL == "" {
					h.Bot.Send(tgbotapi.NewMessage(chatID, "❌ Failed to shorten URL."))
					continue
				}
				msg := tgbotapi.NewMessage(chatID, "✅ Shortened URL: "+shortURL)
				h.Bot.Send(msg)

			default:
				msg := tgbotapi.NewMessage(chatID, "❓ I don't understand. Use the button or type /start.")
				h.Bot.Send(msg)
			}
		}
	}
}

func (h *BotHandler) shortenURL(originalURL string) (string, error) {
	requestBody := strings.NewReader(`{"Url": "` + originalURL + `"}`)
	resp, err := http.Post("http://"+models.Config.HostName+":"+models.Config.Port+"/short", "application/json", requestBody)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", err
	}

	result := models.Respons{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	return result.Url, nil
}
