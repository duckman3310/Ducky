package downloader

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
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

// скачивает аудио дорожку в формате m4a из видео на ютуб по ссылке
func Download(input, outputTemplate string) error {
	args := []string{
		"--newline",
		"--js-runtimes", "node:/usr/bin/nodejs",
		"-f", "ba",
		"-x",
		"--audio-format", "m4a",
		"--audio-quality", "192K",
		"-o", outputTemplate,
		"--write-thumbnail",
		"--embed-thumbnail",
		"--convert-thumbnails", "jpg",
		"--postprocessor-args", "ExtractAudio:-af volume=-1.0dB",
	}

	// парсинг входных аргументов
	var links []string
	for _, arg := range strings.Fields(input) {

		if arg == "" {
			continue
		}

		switch {
		case strings.HasPrefix(arg, "http://"), strings.HasPrefix(arg, "https://"):
			links = append(links, arg)

		case strings.HasPrefix(arg, "*"):
			args = append(args, "--download-sections", arg)
		}
	}

	if len(links) == 0 {
		return errors.New("отсутствует ссылка")
	}

	// сборка вызова
	args = append(args, links...)
	cmd := exec.Command("yt-dlp", args...)

	pipeReader, pipeWriter := io.Pipe()
	cmd.Stderr = pipeWriter
	// 	cmd.Stdout = pipeWriter

	// вызов
	if err := cmd.Start(); err != nil {
		_ = pipeWriter.Close()
		return fmt.Errorf("запуск yt-dlp: %w", err)
	}

	go func() {
		_ = pipeWriter.CloseWithError(cmd.Wait())
	}()

	// запись ошибок
	scanner := bufio.NewScanner(pipeReader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	stderrLines := []string{}

	for scanner.Scan() {
		stderrLines = append(stderrLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("yt-dlp: %w (stderr: %s)", err, strings.Join(stderrLines, "\n"))
	}

	return nil
}
