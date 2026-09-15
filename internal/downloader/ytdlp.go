package downloader

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
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
func YtdlpDownload(outputTemplate, url, quality, downloadSections string) ([]string, error) {

	// создаем пустой лог
	downloadLog := []string{}

	var cmd *exec.Cmd

	// аргументы вызова
	args := []string{
		"--newline",
		"--js-runtimes", "node",
		"-f", "ba",
		"-x",
		"--audio-format", "m4a",
		"-o", outputTemplate,
		url,

		// обложка
		"--write-thumbnail",           // Скачать обложку в отдельный файл (я уже и забыл зачем)
		"--embed-thumbnail",           // Вшить внутрь m4a
		"--convert-thumbnails", "jpg", // Формат JPG для совместимости

		// запас в 1 децибел, при конвертации пики
		// могут выходить на 0 db что приводит к хрипу/скрипу
		"--postprocessor-args", "ExtractAudio:-af volume=-1.0dB",
	}

	// при конвертации .webm(opus) -> .m4a(aac) изза разницы
	// в эффективности сжатия часть деталей теряется,
	// решается это волшебным "--audio-quality", "0",
	// ffmpeg сохраняет звук один в один (даже артефакты opus)
	// раздуювая битрейт до ~400 Kbs

	// если средний вес песни в 8-10мб не устраивает, битрейт можно залочить
	// здесь это 192K (192 Kbs), что тоже вполне неплохо

	if quality == "max" {
		args = append(args, "--audio-quality", "0")
	} else {
		args = append(args, "--audio-quality", "192K")
	}

	if downloadSections != "" {
		args = append(args, "--download-sections", downloadSections) // формат *12:34-12:34
	}

	// использовать куки если они есть
	if _, err := os.Stat("cookies.txt"); err == nil {
		args = append(args, "--cookies", "cookies.txt")
	}

	// финальная сборка
	cmd = exec.Command("yt-dlp", args...)

	pipeReader, pipeWriter := io.Pipe()

	// подключаем врайтер к выходам запускаемой проги
	cmd.Stdout = pipeWriter
	cmd.Stderr = pipeWriter

	// вызов yt-dlp
	if err := cmd.Start(); err != nil {
		_ = pipeWriter.Close()
		return nil, err
	}

	// дожидаемся закрытия программы в отдельной горутине
	go func() {
		err := cmd.Wait()
		_ = pipeWriter.CloseWithError(err)
	}()

	// Создаем буфер для вывода yt-dlp из трубы, сканер проще говоря
	scanner := bufio.NewScanner(pipeReader)

	// сканируем все что выдает yt-dlp
	for scanner.Scan() {
		line := scanner.Text()

		downloadLog = append(downloadLog, line)

		log.Println(Yellow + "yt-dlp: " + Reset + line)
	}

	// Если yt-dlp завершился с ошибкой, CloseWithError передаст её в scanner.Err()
	if err := scanner.Err(); err != nil {
		return downloadLog, err
	}

	return downloadLog, nil
}

func ParseYtDlpError(downloadLog []string) string {
	var rawErr string

	// Ищем строку, где yt-dlp пишет конкретную ошибку
	for _, line := range downloadLog {
		if strings.Contains(line, "ERROR:") {
			rawErr = line
			break
		}
	}

	// Если в логе не нашли явного ERROR:, берем последнюю непустую строку
	if rawErr == "" && len(downloadLog) > 0 {
		rawErr = downloadLog[len(downloadLog)-1]
	}

	// Парсим ошибки yt-dlp и превращаем в "объяснялку" для чата
	switch {
	case strings.Contains(rawErr, "Sign in to confirm your age"):
		return "Видео с ограничением 18+"
	case strings.Contains(rawErr, "Video unavailable"):
		return "Видео недоступно"
	case strings.Contains(rawErr, "Private video"):
		return "Это приватное видео"
	case strings.Contains(rawErr, "is a live stream"):
		return "Нельзя скачивать незавершенные прямые трансляции"
	case strings.Contains(rawErr, "Incomplete YouTube ID"):
		return "Некорректная ссылка на видео"
	default:
		if rawErr != "" {
			return fmt.Sprintf("Ошибка yt-dlp: %s", rawErr)
		}
		return "Произошла неизвестная ошибка при скачивании"
	}
}
