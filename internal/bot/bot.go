package bot

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"tgbot/internal/downloader"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	Log = Green + "[LOG] " + Reset
	Err = Red + "[ERR] " + Reset
	Wrn = Yellow + "[WRN] " + Reset
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
		var downloadSections string

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

func downloadAndSend(bot *tgbotapi.BotAPI, chatID int64, userMsgID int, url string, showlog bool, quality string, downloadSections string) {

	// таймер
	startTime := time.Now()

	// ID текущего скачивания для имени файла
	reqID := strconv.FormatInt(time.Now().UnixNano(), 36)

	// добавляем reqID к логам
	var (
		Log = reqID + Green + " [LOG]   " + Reset
		Err = reqID + Red + " [ERROR] " + Reset
		Wrn = reqID + Yellow + " [WARN]  " + Reset
	)

	// ожидаемое имя файлов
	outputTemplate := fmt.Sprintf("track_%d_%s_%%(title)s.%%(ext)s", chatID, reqID)

	log.Printf("%sполучена ссылка: %s", Log, url)

	// удаляем сообщение со ссылкой от пользователя
	if _, err := bot.Request(tgbotapi.NewDeleteMessage(chatID, userMsgID)); err != nil {
		log.Printf("%sНе удалось отправить запрос на удаление отправленной пользователем ссылки: %v", Err, err)
	}

	// отправляем statusMsg
	statusMsg, err := bot.Send(tgbotapi.NewMessage(chatID, "Скачивается.."))
	if err != nil {
		log.Printf("%sНе удалось отправить statusMsg: %v", Err, err)
	}

	defer func() {
		log.Print(Log + "Удаление файлов")

		files, err := filepath.Glob(fmt.Sprintf("track_%d_%s_*", chatID, reqID))
		if err != nil {
			log.Printf("%sНе удалось получить список файлов для удаления: %v", Err, err)
		}

		for _, file := range files {
			if err = os.Remove(file); err != nil {
				log.Printf("%sНе удалось удалить файл %s: %v", Err, file, err)
			}
		}

		// Удаляем временный статус
		if _, err = bot.Request(tgbotapi.NewDeleteMessage(chatID, statusMsg.MessageID)); err != nil {
			log.Printf("%sНе удалось отправить запрос на удаление statusMsg: %v", Err, err)
		}

		// итоговое время
		endTime := time.Since(startTime).Seconds()
		log.Printf("%sЗавершено, %.1fc", Log, endTime)
	}()

	// ------------------- Вызов yt-dlp -------------------

	log.Print(Log + "Вызов yt-dlp")
	downloadlog, err := downloader.YtdlpDownload(outputTemplate, url, quality, downloadSections)

	// отправка лога
	if showlog {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("заняло %.1fc,\nкачество: %s\nreqID: %s\n```\n%s\n```", time.Since(startTime).Seconds(), quality, reqID, strings.Join(downloadlog, "\n")))
		msg.ParseMode = tgbotapi.ModeMarkdown
		if _, err = bot.Send(msg); err != nil {
			log.Printf("%sНе удалось отправить лог скачивания: %v", Err, err)
		}
	}

	// если yt-dlp накрылся
	if err != nil {
		errReason := downloader.ParseYtDlpError(downloadlog)

		log.Printf("%sНе удалось скачать: %s", Err, errReason)

		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Не удалось скачать, %s\n```Скопировать\n%s```", errReason, url))
		msg.ParseMode = tgbotapi.ModeMarkdown

		if _, err = bot.Send(msg); err != nil {
			log.Printf("%sНе удалось отправить уведомление об ошибке скачивания: %v", Err, err)
		}
		return
	}

	// ------------------- Работа со скачанными файлами -------------------

	// получаем список скачанных файлов
	downloadFiles, err := filepath.Glob(fmt.Sprintf("track_%d_%s_*", chatID, reqID))
	if err != nil {
		log.Printf("%sНе удалось получить список скачанных файлов: %v", Err, err)
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
		log.Print(Wrn + "Превью файл не найден")
		if audioPath == "" {
			log.Print(Err + "Аудио файл не найден")
			if _, err = bot.Send(tgbotapi.NewEditMessageText(chatID, statusMsg.MessageID, "аудио файл утерян")); err != nil {
				log.Printf("%sНе удалось отправить уведомление об утерянном аудио файле: %v", Err, err)
			}
			return
		}
	}

	// расширение
	actualExt := filepath.Ext(audioPath)

	// размер
	var fileSizeMB float64
	if fileInfo, err := os.Stat(audioPath); err == nil {
		fileSizeMB = float64(fileInfo.Size()) / 1024 / 1024
	}

	// Собираем имя файла для Телеграма
	cleanName := filepath.Base(audioPath)
	cleanName = strings.TrimPrefix(cleanName, fmt.Sprintf("track_%d_%s_", chatID, reqID))
	cleanName = strings.TrimSuffix(cleanName, actualExt)

	log.Printf("%sСкачан файл: '%s%s' (%.2f MB)", Log, cleanName, actualExt, fileSizeMB)

	// Меняем статус в Телеграме
	if _, err = bot.Send(tgbotapi.NewEditMessageText(chatID, statusMsg.MessageID, fmt.Sprintf("%s Отправляется..", cleanName))); err != nil {
		log.Printf("%sНе удалось отправить уведомление о скачанном файле: %v", Err, err)
	}

	// Создаем объект аудио
	audioFile := tgbotapi.NewAudio(chatID, tgbotapi.FilePath(audioPath))
	audioFile.Title = cleanName
	audioFile.Performer = ""
	audioFile.Thumb = tgbotapi.FilePath(previewPath)

	// Отправляем файл
	if _, err := bot.Send(audioFile); err != nil {
		log.Printf("%sТелеграм отклонил отправку файла: %v", Err, err)
		if _, err = bot.Send(tgbotapi.NewEditMessageText(chatID, statusMsg.MessageID, "Телеграм отклонил отправку файла")); err != nil {
			log.Printf("%sНе удалось отправить уведомление об отклонении отправки файла: %v", Err, err)
		}
		return
	}

	log.Print(Log + "Файл отправлен")
}
