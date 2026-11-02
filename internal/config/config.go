package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Token       string `toml:"bot_token"`
	InfoText    string `toml:"info_text"`
	SongQuality string `toml:"song_quality"`
}

// Load загружает конфиг или создает дефолтный
func Load(path string) (*Config, error) {
	var cfg Config

	_, err := toml.DecodeFile("./"+path, &cfg)

	if errors.Is(err, os.ErrNotExist) {
		defaultCfg := `
			# Токен бота от @BotFather
			token = "Your bot token"

			# Текст ответа для /info
			info_text = ""

			# Качество звука: "default" или "max"
			song_quality = "default"`

		if wErr := os.WriteFile(path, []byte(defaultCfg), 0644); wErr != nil {
			return nil, fmt.Errorf("не удалось создать дефолтный конфиг: %w", wErr)
		}
		return nil, fmt.Errorf("создан новый %s, заполните его и перезапустите бота", path)
	}

	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать конфиг: %w", err)
	}

	return &cfg, nil
}
