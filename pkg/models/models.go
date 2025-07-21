package models

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type RequestData struct {
	Url string `json:"url"`
}

type Respons struct {
	Url string `json:"url"`
}

type Url struct {
	Telegram_id int64  `json:"Telegram_id"`
	Hash        string `json:"Hash"`
	Url         string `json:"Url"`
}

type LogAction struct {
	Telegram_id int64  `json:"Telegram_id"`
	Action      string `json:"Action"`
}

type LogError struct {
	Telegram_id int64  `json:"Telegram_id"`
	Error       string `json:"Error"`
	Error_code  string `json:"Error_code"`
}

type Users struct {
	Telegram_id int64
	Nick_Name   string
}

type TelegramBot interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	GetUpdatesChan(tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
}

var SqlRequests = map[string]string{
	"users_info": `
	CREATE TABLE IF NOT EXISTS users_info (
		uuid uuid DEFAULT gen_random_uuid() PRIMARY KEY,
		"Nick_Name" TEXT NOT NULL,
		"Telegram_id" BIGINT NOT NULL,
		created_at TIMESTAMP DEFAULT now()
	);
	`,
	"urls": `
		CREATE TABLE IF NOT EXISTS urls (
			uuid uuid DEFAULT gen_random_uuid() PRIMARY KEY,
			"Telegram_id" BIGINT NOT NULL,
			"Hash" TEXT NOT NULL,
			"Url" TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT now()
		);
	`,
	"log_action": `
		CREATE TABLE IF NOT EXISTS log_action (
			uuid uuid DEFAULT gen_random_uuid() PRIMARY KEY,
			"Telegram_id" BIGINT NOT NULL,
			"Action" TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT now()
		);
	`,
	"log_error": `
		CREATE TABLE IF NOT EXISTS log_error (
			uuid uuid DEFAULT gen_random_uuid() PRIMARY KEY,
			"Telegram_id" BIGINT NOT NULL,
			"Error" TEXT NOT NULL,
			"Error_code" TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT now()
		);
	`}

type SupabaseResponse []Url

var Config ConfigStruct

var Protocol string
