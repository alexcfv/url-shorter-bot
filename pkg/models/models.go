package models

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type RequestData struct {
	Url string `json:"url"`
}

type Respons struct {
	Url string `json:"url"`
}

type Url struct {
	Hash string
	Url  string
}

type Users struct {
	Telegram_id string
	Nick_Name   string
}

type TelegramBot interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	GetUpdatesChan(tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
}

var SqlRequests = map[string]string{
	"urls": `
		CREATE TABLE IF NOT EXISTS urls (
			uuid uuid DEFAULT gen_random_uuid() PRIMARY KEY,
			"Hash" TEXT NOT NULL,
			"Url" TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT now()
		);
	`,
	"users": `
		CREATE TABLE IF NOT EXISTS users (
			uuid uuid DEFAULT gen_random_uuid() PRIMARY KEY,
			"Nick_Name" TEXT NOT NULL,
			"Telegram_id" TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT now()
		);
	`,
	"log_action": `
		CREATE TABLE IF NOT EXISTS log_action (
			uuid uuid DEFAULT gen_random_uuid() PRIMARY KEY,
			"Error" TEXT NOT NULL,
			"Error_code" TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT now()
		);
	`,
	"log_error": `
		CREATE TABLE IF NOT EXISTS log_error (
			uuid uuid DEFAULT gen_random_uuid() PRIMARY KEY,
			"Nick_Name" TEXT NOT NULL,
			"Telegram_id" TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT now()
		);
	`}

type SupabaseResponse []Url

var Config ConfigStruct

var Protocol string
