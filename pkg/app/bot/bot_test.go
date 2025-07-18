package bot

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"url-shorter-bot/pkg/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type mockSupabase struct {
	data map[string]string
}

func (m *mockSupabase) Get(table string, target map[string]string) ([]byte, error) {
	key, ok := target["Hash"]
	if !ok {
		return nil, fmt.Errorf("missing filter key 'hash'")
	}

	val, exists := m.data[key]
	if !exists {
		return nil, fmt.Errorf("no record found for hash: %s", key)
	}

	result := map[string]string{
		"Hash":         key,
		"original_url": val,
	}

	return json.Marshal(result)
}

func (m *mockSupabase) Insert(table string, data interface{}) ([]byte, error) {
	if m.data == nil {
		m.data = map[string]string{}
	}
	u, ok := data.(models.Url)
	if !ok {
		return nil, fmt.Errorf("invalid format")
	}
	m.data[u.Hash] = u.Url
	return json.Marshal(u)
}

func (m *mockSupabase) Delete(table string, filter string) ([]byte, error) {
	delete(m.data, "12345")
	return []byte(`{}`), nil
}

type mockBotAPI struct {
	sentMessages []tgbotapi.Chattable
}

func (m *mockBotAPI) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	m.sentMessages = append(m.sentMessages, c)
	return tgbotapi.Message{}, nil
}

func (m *mockBotAPI) GetUpdatesChan(u tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	ch := make(chan tgbotapi.Update)
	close(ch)
	return tgbotapi.UpdatesChannel(ch)
}

func TestHandleMessages(t *testing.T) {
	tests := []struct {
		name           string
		initialState   string
		inputText      string
		mockHTTPStatus int
		mockHTTPBody   string
		expectedReply  string
	}{
		{
			name:          "start command",
			inputText:     "/start",
			expectedReply: "👋 Welcome! Click the button below to shorten a URL.",
		},
		{
			name:          "Shorten URL button",
			inputText:     "Shorten URL",
			expectedReply: "Please send the URL you want to shorten.",
		},
		{
			name:           "valid URL shorten",
			inputText:      "https://google.com",
			initialState:   "awaiting_url",
			mockHTTPStatus: http.StatusOK,
			mockHTTPBody:   `{"Url": "http://short.ly/abc123"}`,
			expectedReply:  "✅ Shortened URL: http://short.ly/abc123",
		},
		{
			name:           "rate limited",
			inputText:      "https://too-many.com",
			initialState:   "awaiting_url",
			mockHTTPStatus: http.StatusTooManyRequests,
			expectedReply:  "Too Many Request",
		},
		{
			name:          "unknown input",
			inputText:     "random input",
			expectedReply: "❓ I don't understand. Use the button or type /start.",
		},
	}

	originalHost := models.Config.HostName
	originalPort := models.Config.Port
	defer func() {
		models.Config.HostName = originalHost
		models.Config.Port = originalPort
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBot := &mockBotAPI{}
			state := NewStateStore()
			db := &mockSupabase{}

			if tt.initialState != "" {
				state.Set(12345, tt.initialState)
			}

			if tt.mockHTTPStatus != 0 {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.mockHTTPStatus)
					w.Write([]byte(tt.mockHTTPBody))
				}))
				defer server.Close()

				hostPort := strings.TrimPrefix(server.URL, "http://")
				host, port, err := net.SplitHostPort(hostPort)
				if err != nil {
					t.Fatalf("failed to parse test server URL: %v", err)
				}
				models.Config.HostName = host
				models.Config.Port = port
			}

			handler := &BotHandler{
				Bot:   mockBot,
				State: state,
				Db:    db,
			}

			update := tgbotapi.Update{
				Message: &tgbotapi.Message{
					Text: tt.inputText,
					Chat: &tgbotapi.Chat{ID: 12345},
					From: &tgbotapi.User{ID: 999},
				},
			}

			processMessage(handler, update)

			if len(mockBot.sentMessages) == 0 {
				t.Fatal("No messages were sent")
			}

			last := mockBot.sentMessages[len(mockBot.sentMessages)-1].(tgbotapi.MessageConfig)
			if !strings.Contains(last.Text, tt.expectedReply) {
				t.Errorf("Expected reply to contain %q, got %q", tt.expectedReply, last.Text)
			}
		})
	}
}

func TestRateLimitBehavior(t *testing.T) {
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount <= 2 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"Url": "http://short.ly/success"}`))
		} else {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("Too Many Request"))
		}
	}))
	defer server.Close()

	parts := strings.Split(strings.TrimPrefix(server.URL, "http://"), ":")
	models.Config.HostName = parts[0]
	models.Config.Port = parts[1]

	state := NewStateStore()
	mockBot := &mockBotAPI{}
	handler := &BotHandler{
		Bot:   mockBot,
		State: state,
	}

	chatID := int64(777)

	for i := 1; i <= 3; i++ {
		state.Set(chatID, "awaiting_url")
		update := tgbotapi.Update{
			Message: &tgbotapi.Message{
				Text: "https://spam.com",
				Chat: &tgbotapi.Chat{ID: chatID},
				From: &tgbotapi.User{ID: 999},
			},
		}
		processMessage(handler, update)
	}

	if len(mockBot.sentMessages) != 3 {
		t.Fatalf("expected 3 responses, got %d", len(mockBot.sentMessages))
	}

	last := mockBot.sentMessages[2].(tgbotapi.MessageConfig)
	if !strings.Contains(last.Text, "Too Many Request") {
		t.Errorf("expected last reply to be rate limit, got: %q", last.Text)
	}
}

// Извлечённая логика из Run(), чтобы тестировать отдельно
func processMessage(h *BotHandler, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	text := update.Message.Text
	telegramID := update.Message.From.ID

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
		shortURL, err := h.shortenURL(text, telegramID)
		if err != nil || shortURL == "" {
			h.Bot.Send(tgbotapi.NewMessage(chatID, "❌ Failed to shorten URL."))
			return
		}
		if shortURL == "Too Many Request" {
			h.Bot.Send(tgbotapi.NewMessage(chatID, "Too Many Request"))
			return
		}
		h.Bot.Send(tgbotapi.NewMessage(chatID, "✅ Shortened URL: "+shortURL))

	default:
		msg := tgbotapi.NewMessage(chatID, "❓ I don't understand. Use the button or type /start.")
		h.Bot.Send(msg)
	}
}
