package bot

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (bot *Bot) Run() {

	bot.API.Debug = false
	updates := bot.API.GetUpdatesChan(tgbotapi.UpdateConfig{Timeout: 60})

	if _, err := bot.API.Request(tgbotapi.SetMyCommandsConfig{Commands: bot.Commands}); err != nil {
		logErr("Не удалось отправить команды", err)
	}

	logBot(fmt.Sprintf("Ducky запущен, cобран из коммита: %s", GetVersion()))

	for update := range updates {

		// если ничего нету скип
		if update.Message == nil || update.Message.Text == "" {
			continue
		}

		messageText := update.Message.Text
		chatID := update.Message.Chat.ID
		messageID := update.Message.MessageID

		if update.Message.IsCommand() {
			bot.replyCommand(messageText, chatID)
			continue
		}

		if strings.HasPrefix(messageText, "http://") || strings.HasPrefix(messageText, "https://") {
			go bot.downloadAndSend(chatID, messageID, messageText)
		}

	}
}
