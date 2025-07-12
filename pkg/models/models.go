package models

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type RequestData struct {
	Url string `json:"url"`
}

type Respons struct {
	Url string `json:"Url"`
}

type Url struct {
	Hash string
	Url  string
}

type TelegramBot interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	GetUpdatesChan(tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
}

type SupabaseResponse []Url

var Config ConfigStruct

var Protocol string
