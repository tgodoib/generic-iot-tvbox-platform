package monitor

import (
	"github.com/go-telegram/bot"
)

type TelegramAlert struct {
	bot    *bot.Bot
	chatID string
}

func NewTelegramAlert(chatID string, token string) TelegramAlert {
	b, err := bot.New(token)
	if err != nil {
		panic(err)
	}

	return TelegramAlert{
		bot:    b,
		chatID: chatID,
	}
}
