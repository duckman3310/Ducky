package bot

import (
	"ducky/internal/downloader"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// скачивает и отправляет в чат аудио файл
func (bot *Bot) downloadAndSend(chatID int64, messageID int, messageText string) {

	logBot(fmt.Sprintf("получено: %s", messageText), chatID)

	startTime := time.Now() // таймер

	reqID := strconv.FormatInt(time.Now().UnixNano(), 36)                // ID текущего скачивания, для имени файла
	outputTemplate := fmt.Sprintf("track_%s_%%(title)s.%%(ext)s", reqID) // ожидаемое имя файлов

	statusMsg, err := bot.API.Send(tgbotapi.NewMessage(chatID, "Скачивается.."))
	if err != nil {
		logErr("Не удалось отправить statusMsg", err, chatID)
	}

	if _, err := bot.API.Request(tgbotapi.NewDeleteMessage(chatID, messageID)); err != nil {
		logErr("Не удалось удалить сообщение пользователя", err, chatID)
	}

	// удаление скачаных файлов в конце
	defer func() {

		files, err := filepath.Glob(fmt.Sprintf("track_%s_*", reqID))
		if err != nil {
			logErr("Не удалось получить список файлов на удаление ", err, chatID)
		}

		for _, file := range files {
			if err = os.Remove(file); err != nil {
				logErr("Не удалось удалить файл "+file, err, chatID)
			}
		}
	}()

	// скачивание
	if err := downloader.Download(messageText, outputTemplate); err != nil {
		logErr("Не удалось скачать", err, chatID)

		errMsg := tgbotapi.NewEditMessageText(chatID, statusMsg.MessageID, fmt.Sprintf("Не удалось скачать, %s ", err))

		if _, err = bot.API.Send(errMsg); err != nil {
			logErr("Не удалось отправить в чат ошибку скачивания", err, chatID)
		}

		return
	}

	// получаем пути скачанных файлов
	downloadFiles, err := filepath.Glob(fmt.Sprintf("track_%s_*", reqID))
	if err != nil {
		logErr("Не удалось получить пути скачанных файлов", err, chatID)
		return
	}

	var audioPath string
	var previewPath string

	// получаем имена .m4a и .jpg файлов
	for _, file := range downloadFiles {
		if strings.HasSuffix(file, ".m4a") {
			audioPath = file
		} else if strings.HasSuffix(file, ".jpg") {
			previewPath = file
		}
	}

	// если файлы затерялись
	if previewPath == "" {
		logWarn("Превью файл не найден", chatID)
	}

	if audioPath == "" {
		logErr("Аудио файл не найден", nil, chatID)

		if _, err = bot.API.Send(tgbotapi.NewEditMessageText(chatID, statusMsg.MessageID, "Аудио файл на отправку не найден")); err != nil {
			logErr("Не удалось отправить уведомление об утерянном аудио файле", err, chatID)
		}

		return
	}

	// Собираем имя файла для Телеграма
	cleanName := filepath.Base(audioPath)
	cleanName = strings.TrimPrefix(cleanName, fmt.Sprintf("track_%s_", reqID))
	cleanName = strings.TrimSuffix(cleanName, ".m4a")

	// Создаем объект аудио
	audioObj := tgbotapi.NewAudio(chatID, tgbotapi.FilePath(audioPath))
	audioObj.Title = cleanName
	audioObj.Performer = ""
	audioObj.Thumb = tgbotapi.FilePath(previewPath)

	// Меняем статус в Телеграме
	if _, err = bot.API.Send(tgbotapi.NewEditMessageText(chatID, statusMsg.MessageID, fmt.Sprintf("%s Отправляется..", cleanName))); err != nil {
		logErr("Не удалось отправить уведомление о скачанном файле", err, chatID)
	}

	// Отправляем файл
	if _, err := bot.API.Send(audioObj); err != nil {

		logErr("Телеграм отклонил отправку файла", err, chatID)

		if _, err = bot.API.Send(tgbotapi.NewEditMessageText(chatID, statusMsg.MessageID, "Телеграм отклонил отправку файла")); err != nil {
			logErr("Не удалось отправить уведомление об отклонении отправки файла", err, chatID)
		}

		return
	}

	if _, err = bot.API.Request(tgbotapi.NewDeleteMessage(chatID, statusMsg.MessageID)); err != nil {
		logErr("Не удалось отправить запрос на удаление statusMsg ", err, chatID)
	}

	endTime := time.Since(startTime).Seconds()
	logBot(fmt.Sprintf("Скачивание и отправка завершены (%.1f c)", endTime), chatID)
}
