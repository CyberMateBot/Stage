package bot

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api *tgbotapi.BotAPI
}

// New creates Telegram bot when TELEGRAM_BOT_ENABLED=true and TELEGRAM_BOT_TOKEN is set.
func New() (*Bot, error) {
	if !botEnabled() {
		slog.Info("telegram bot disabled (TELEGRAM_BOT_ENABLED=false)")
		return &Bot{}, nil
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		slog.Info("telegram bot disabled (TELEGRAM_BOT_TOKEN is empty)")
		return &Bot{}, nil
	}

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	expected := os.Getenv("TELEGRAM_BOT_USERNAME")
	if expected != "" && api.Self.UserName != expected {
		return nil, fmt.Errorf("bot username mismatch: got @%s, want @%s", api.Self.UserName, expected)
	}

	slog.Info("telegram bot connected", slog.String("username", "@"+api.Self.UserName))
	return &Bot{api: api}, nil
}

func botEnabled() bool {
	v := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_ENABLED"))
	if v == "" {
		return true
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// StartPolling starts handling /start command.
func (b *Bot) StartPolling(ctx context.Context) {
	if b.api == nil {
		return
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	updates := b.api.GetUpdatesChan(u)
	for {
		select {
		case <-ctx.Done():
			return
		case upd := <-updates:
			if upd.Message != nil && upd.Message.IsCommand() && upd.Message.Command() == "start" {
				name := upd.Message.From.FirstName
				msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "Добро пожаловать в CyberMate, "+name+"💚\n– Получайте рекламные задания...")
				if _, err := b.api.Send(msg); err != nil {
					slog.ErrorContext(ctx, "failed to send telegram message", slog.Any("error", err))
				}
			}
		}
	}
}

// SendMessage sends a text with optional inline buttons.
func (b *Bot) SendMessage(telegramID int64, text string, buttons []tgbotapi.InlineKeyboardButton) error {
	if b.api == nil {
		return nil
	}
	msg := tgbotapi.NewMessage(telegramID, text)
	if len(buttons) > 0 {
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons)
	}
	_, err := b.api.Send(msg)
	return err
}
