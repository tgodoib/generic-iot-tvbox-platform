package platforms

import (
	"context"
	"fmt"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Telegram struct {
	bot *bot.Bot
}

type TelegramRecipient struct {
	*Telegram
	chatID string
}

func NewTelegram(token string) Telegram {
	b, err := bot.New(token)
	if err != nil {
		panic(err)
	}

	return Telegram{
		bot: b,
	}
}

func (t *Telegram) NewRecipient(chatID string) *TelegramRecipient {
	return &TelegramRecipient{
		Telegram: t,
		chatID:   chatID,
	}
}

func (t *TelegramRecipient) Send(title string, body string, deviceId string) error {
	fmt.Println("Sending msg")

	msg := fmt.Sprintf(""+
		"<b>%s</b>\n"+
		"<pre>Device: %s</pre>\n"+
		"<pre>%s</pre>",
		title,
		deviceId,
		body,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    t.chatID,
		Text:      msg,
		ParseMode: models.ParseModeHTML,
	})

	return err
}
