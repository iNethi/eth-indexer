package telegram

import (
	"context"

	"github.com/mr-linch/go-tg"
)

type (
	TelegramOpts struct {
		BotToken            string
		NotificationChannel int64
	}

	Telegram struct {
		client               *tg.Client
		notificaationChannel int64
	}
)

const (
	NOTIFY_LOW_BALANCE_ON_GAS_FAUCET = `
		Gas faucet balance is low. Top is required soon!`
)

func New(o TelegramOpts) *Telegram {
	return &Telegram{
		client:               tg.New(o.BotToken, nil),
		notificaationChannel: o.NotificationChannel,
	}
}

func (t *Telegram) Notify(ctx context.Context, message string) error {
	_, err := t.client.SendMessage(tg.ChatID(t.notificaationChannel), message).Do(ctx)
	return err

}
