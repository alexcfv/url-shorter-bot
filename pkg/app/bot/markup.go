package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func UrlShortenKeyboard() tgbotapi.ReplyKeyboardMarkup {
	button := tgbotapi.NewKeyboardButton("Shorten URL")
	keyboard := tgbotapi.NewReplyKeyboard(
		[]tgbotapi.KeyboardButton{
			button,
		},
	)
	keyboard.OneTimeKeyboard = true
	keyboard.ResizeKeyboard = true
	return keyboard
}
