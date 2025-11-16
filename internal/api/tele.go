package api

import (
	"context"
	"log/slog"
	"strconv"
	"sync"
	"unicode"

	tele "gopkg.in/telebot.v3"
)

func sendMessage(b *tele.Bot, chatID, message string) {
	i, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		slog.Error("Error parsing chatID", "err", err)
	}
	chat := &tele.Chat{ID: i}

	// Преобразуем текст в rune slice для корректной работы с символами
	textRunes := []rune(message)

	// Разбиваем текст на части по 3000 символов
	const maxLen = 3000
	var parts []string

	for len(textRunes) > maxLen {
		// Ищем последний пробел перед лимитом
		lastSpace := -1
		for i := maxLen - 1; i >= 0; i-- {
			if unicode.IsSpace(textRunes[i]) {
				lastSpace = i
				break
			}
		}
		if lastSpace > 0 {
			parts = append(parts, string(textRunes[:lastSpace]))
			textRunes = textRunes[lastSpace+1:]
		} else {
			parts = append(parts, string(textRunes[:maxLen]))
			textRunes = textRunes[maxLen:]
		}
	}

	if len(textRunes) > 0 {
		parts = append(parts, string(textRunes))
	}

	// Отправляем каждую часть
	for _, part := range parts {
		_, err = b.Send(chat, part)
		if err != nil {
			slog.Error("Telegram error", "err", err)
			return
		}
	}
}

func TgSendWorker(ctx context.Context, b *tele.Bot, telegramChatID string, ch <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	slog.Info("tgSendWorker started")
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			sendMessage(b, telegramChatID, msg)
		}
	}
}
