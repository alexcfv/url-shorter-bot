package bot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"url-shorter-bot/pkg/database"
	"url-shorter-bot/pkg/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotHandler struct {
	Bot   models.TelegramBot
	State *StateStore
	Db    database.SupabaseClient
}

func NewBotHandler(token string, state *StateStore, db database.SupabaseClient) (*BotHandler, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &BotHandler{Bot: bot, State: state, Db: db}, nil
}

func (h *BotHandler) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := h.Bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			chatID := update.Message.Chat.ID
			text := update.Message.Text
			telegramID := update.Message.From.ID
			username := update.Message.From.UserName

			switch {
			case text == "/start":
				_, err := h.Db.Insert("users", models.Users{
					Telegram_id: telegramID,
					Nick_Name:   username,
				})
				if err != nil {
					fmt.Println("Error to write user into users table")
				}

				msg := tgbotapi.NewMessage(chatID, "👋 Welcome! Click the button below to shorten a URL.")
				msg.ReplyMarkup = UrlShortenKeyboard()
				h.Bot.Send(msg)

			case text == "Shorten URL":
				h.State.Set(chatID, "awaiting_url")
				msg := tgbotapi.NewMessage(chatID, "Please send the URL you want to shorten.")
				h.Bot.Send(msg)

			case h.State.Get(chatID) == "awaiting_url":
				h.State.Clear(chatID)
				shortURL, err := h.shortenURL(text, telegramID)
				if err != nil || shortURL == "" {
					h.Bot.Send(tgbotapi.NewMessage(chatID, "❌ Failed to shorten URL."))
					continue
				}
				if shortURL == "Too Many Request" {
					msg := tgbotapi.NewMessage(chatID, shortURL)
					h.Bot.Send(msg)
				} else {
					msg := tgbotapi.NewMessage(chatID, "✅ Shortened URL: "+shortURL)
					h.Bot.Send(msg)
				}

			default:
				msg := tgbotapi.NewMessage(chatID, "❓ I don't understand. Use the button or type /start.")
				h.Bot.Send(msg)
			}
		}
	}
}

func (h *BotHandler) shortenURL(originalURL string, telegramID int64) (string, error) {
	client := &http.Client{}
	requestBody := strings.NewReader(`{"Url": "` + originalURL + `"}`)

	req, _ := http.NewRequest("POST", "http://"+models.Config.HostName+":"+models.Config.Port+"/short", requestBody)
	req.Header.Set("X-Telegram-ID", strconv.FormatInt(telegramID, 10))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusTooManyRequests {
		return "", err
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return "Too Many Request", nil
	}

	result := models.Respons{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	return result.Url, nil
}
