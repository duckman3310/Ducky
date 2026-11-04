package bot

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var Commands = []tgbotapi.BotCommand{
	{Command: "info", Description: "Информация о боте"},
	{Command: "logs", Description: "Логи текущего чата"},
	{Command: "help", Description: "Как скачать аудио с видео?"},
}

// отправляет ответ на комманду
func (bot *Bot) replyCommand(command string, chatID int64) {
	switch command {
	case "/info":
		if bot.Cfg.InfoText != "" {
			if _, err := bot.API.Send(tgbotapi.NewMessage(chatID, bot.Cfg.InfoText)); err != nil {
				logErr("Не удалось отправить ответ на команду /info", err, chatID)
			}
		}
	case "/logs":
		logs := tgbotapi.NewMessage(chatID, strings.Join(UserLogs.Read(chatID), "\n"))
		if _, err := bot.API.Send(logs); err != nil {
			logErr("Не удалось отправить ответ на команду /logs", err, chatID)
		}
	case "/help":
		helpMsg := tgbotapi.NewMessage(chatID, "Принимается: \nСыллка (Youtube, VK, TikTok, Instagram и т.д.)\nТаймкoд от и до (*ЧЧ:ММ:СС, 0 - сначала, inf - до конца)\n\n Пример: https://youtu.be/dQw4w9WgXcQ?si=lzgVyvPadCL0DKz7 *00:35-1:20")
		helpMsg.DisableWebPagePreview = true
		if _, err := bot.API.Send(helpMsg); err != nil {
			logErr("Не удалось отправить ответ на команду /help", err, chatID)
		}
	}
}
