package bot

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"tgbot/internal/downloader"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	Log = Green + "[LOG]  " + Reset
	Err = Red + "[ERR]  " + Reset
)

func BotInit() {

	// загружаем конфиг
	botcfg := NewBotCfg()
	if err := botcfg.Load(); err != nil {
		log.Printf(Err+"Не удалось загрузить конфигурацию: %v", err)
		return
	}

	// инициализируем бота
	bot, err := tgbotapi.NewBotAPI(botcfg.Token)
	if err != nil {
		log.Printf(Err+"Не удалось инициализировать бота: %v", err)
		return
	}

	// отправляем команды
	if _, err := bot.Request(commandsConfig); err != nil {
		log.Printf(Err+"Не удалось отправить команды: %v", err)
	}

	// настройка
	bot.Debug = false
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	log.Printf(Log + "Бот запущен")

	for update := range updates {

		// если ничего нету скип
		if update.Message == nil || update.Message.Text == "" {
			continue
		}

		messageText := update.Message.Text
		chatID := update.Message.Chat.ID
		messageID := update.Message.MessageID

		// если команда
		if update.Message.IsCommand() {
			switch update.Message.Text {
			case "/info":
				if botcfg.InfoText != "" {
					if _, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, botcfg.InfoText)); err != nil {
						log.Printf(Red+"Не удалось отправить ответ на команду /info"+Reset+": %v", err)
					}
				}
			}
			continue
		}

		showLog := false
		var link string
		var downloadSections string = "*0-inf"

		// cобираем аргументы
		for _, arg := range strings.Fields(messageText) {

			if strings.HasPrefix(arg, "*") {
				downloadSections = arg
			}

			if strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://") {
				link = arg
			}

			if arg == "-l" {
				showLog = true
			}
		}

		// если чет есть то скачиваем
		if link != "" {
			go downloadAndSend(bot, chatID, messageID, link, showLog, botcfg.SongQuality, downloadSections)
		}
	}
}

// downloadAndSend вызывает yt-dlp с сохранением лога
// для скачивания по ссылке и отправляет файл в чат
func downloadAndSend(bot *tgbotapi.BotAPI, chatID int64, userMsgID int, url string, showlog bool, quality string, downloadSections string) {

	// ставим таймер
	startTime := time.Now()

	log.Printf(Log+"получена ссылка: %s", url)

	// удаляем сообщение с ссылкой от пользователя
	if _, err := bot.Request(tgbotapi.NewDeleteMessage(chatID, userMsgID)); err != nil {
		log.Printf(Err+"Не удалось отправить запрос на удаление отправленой пользователем ссылки: %v", err)
	}

	// создаем и отправляем statusMsg
	statusMsg, err := bot.Send(tgbotapi.NewMessage(chatID, "Скачивается.."))
	if err != nil {
		log.Printf(Err+"Не удалось отправить statusMsg: %s", err)
	}

	// Уникальный ID текущего скачивания для имени файла
	// чтобы параллельные потоки их не путали
	reqID := time.Now().UnixNano()

	// формируем ожидаемое имя файла
	outputTemplate := fmt.Sprintf("track_%d_%d_%%(title)s.%%(ext)s", chatID, reqID) // %%(title)s.%%(ext)s - сюда yt-dlp подставит имя и разширение

	// Вызов yt-dlp и сбор лога его скачивания
	log.Printf(Log + "Вызов yt-dlp: ")
	downloadlog, err := downloader.YtdlpDownload(outputTemplate, url, quality, downloadSections)
	if err != nil {
		log.Printf(Err+"Не удалось скачать файл, %s", err)
		if _, err = bot.Send(tgbotapi.NewEditMessageText(chatID, statusMsg.MessageID, "Не удалось скачать:\n\n"+url)); err != nil {
			log.Printf(Err+"Не удалось отправить уведомление о ошибке скачивания: %v", err)
		}
		return
	}

	// ищем скачанный файл
	audioFiles, err := filepath.Glob(fmt.Sprintf("track_%d_%d_*.m4a", chatID, reqID))
	if err != nil || len(audioFiles) == 0 {
		log.Printf(Err + "Не удалось найти скачанный файл")
		if _, err = bot.Send(tgbotapi.NewEditMessageText(chatID, statusMsg.MessageID, "Не удалось найти скачанный файл")); err != nil {
			log.Printf(Err+"Не удалось отправить уведомление о потеряном файле: %v", err)
		}
		return
	}

	filePath := audioFiles[0]

	// удаляем файл с сервера при завершении функции
	defer func() {
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			log.Printf(Err+"Не удалось удалить файл с сервера: %v", err)
		}
	}()

	actualExt := filepath.Ext(filePath)

	var fileSizeMB float64
	if fileInfo, err := os.Stat(filePath); err == nil {
		fileSizeMB = float64(fileInfo.Size()) / 1024 / 1024
	}

	// Собираем имя файла для телеграмма
	cleanName := filepath.Base(filePath)
	cleanName = strings.TrimPrefix(cleanName, fmt.Sprintf("track_%d_%d_", chatID, reqID))
	cleanName = strings.TrimSuffix(cleanName, actualExt)

	log.Printf(Log+"Скачан файл: '%s%s %.2f' MB", cleanName, actualExt, fileSizeMB)

	// Меняем статус в Телеграме
	if _, err = bot.Send(tgbotapi.NewEditMessageText(chatID, statusMsg.MessageID, fmt.Sprintf("%s Отправляется..", cleanName))); err != nil {
		log.Printf(Err+"Не удалось отправить уведомление о скачанном файле: %v", err)
	}

	// ищем обложку по имени скаченого файла
	var thumbPath string
	possibleThumb := strings.TrimSuffix(filePath, actualExt) + ".jpg"
	if _, err := os.Stat(possibleThumb); err == nil {
		thumbPath = possibleThumb
		defer func(p string) {
			_ = os.Remove(p)
		}(thumbPath)
	}

	// Создаем объект аудио
	audioFile := tgbotapi.NewAudio(chatID, tgbotapi.FilePath(filePath))
	audioFile.Title = cleanName
	audioFile.Performer = ""

	// Прикрепляем обложку
	if thumbPath != "" {
		audioFile.Thumb = tgbotapi.FilePath(thumbPath)
	}

	// Отправляем файл
	if _, err := bot.Send(audioFile); err != nil {
		log.Printf(Err+"Телеграм отклонил отправку файла: %v", err)
		if _, err = bot.Send(tgbotapi.NewEditMessageText(chatID, statusMsg.MessageID, "Телеграм отклонил отправку файла")); err != nil {
			log.Printf(Err+"Не удалось отправить уведомление об отклонении отправки файла: %v", err)
		}
		return
	}

	log.Printf(Log + "Файл отправлен")

	// Удаляем временный статус
	if _, err = bot.Request(tgbotapi.NewDeleteMessage(chatID, statusMsg.MessageID)); err != nil {
		log.Printf(Err+"Не удалось отправить запрос на удаление statusMsg: %v", err)
	}

	// отправка лога скачивания
	if showlog == true {
		downloadTime := time.Since(startTime).Seconds()
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("скачивание и отправка заняло %.1fc,\nкачество: %s\n```\n%s\n```", downloadTime, quality, strings.Join(downloadlog, "\n")))

		msg.ParseMode = tgbotapi.ModeMarkdown // Markdown для красоты

		if _, err = bot.Send(msg); err != nil {
			log.Printf(Err+"Не удалось отправить лог скачивания: %v", err)
		}
	}

	endTime := time.Since(startTime).Seconds()
	log.Printf(Log+"Завершено, %.1fc", endTime)
}
