package bot

import (
	"ducky/internal/config"
	"runtime/debug"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// цвета для терминала
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
)

type Bot struct {
	Cfg *config.Config
	API *tgbotapi.BotAPI

	Commands []tgbotapi.BotCommand
}

var UserLogs LogStore

func New() *Bot {

	cfg, err := config.Load("config.toml")
	if err != nil {
		logErr("Не удалось загрузить конфиг бота", err)
	}

	api, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		logErr("Не удалось инициализировать бота", err)
	}

	return &Bot{
		Cfg: cfg,
		API: api,
		Commands: []tgbotapi.BotCommand{
			{
				Command:     "info",
				Description: "Информация о боте",
			}},
	}
}

func GetVersion() string {

	var version string

	info, ok := debug.ReadBuildInfo()
	if !ok {
		logWarn("Не удалось получить версию бота")
		return "unknown"
	}

	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			version = setting.Value // хеш коммита

		case "vcs.modified":
			if setting.Value == "true" {
				version += " (modified)" // если были изменения после коммита
			}
		}
	}

	return version

}
