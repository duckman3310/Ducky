package downloader

import (
	"bufio"
	"fmt"
	"io"
	"log"
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

	// создаем комманду для скачивания
	if quality == "max" {

		// (скачивает в максимальном качестве)

		// при обычной загрузке во время конвертации
		// какого нибуть .opus (который тг воспринивает как голосовые) в .m4a
		// ffmpeg, изза ограничений в размере часть звука срезается (так называемый true peak),
		// и если вы очень ушастый высокие частоты начинают свистеть/песочить.

		// решается это бустом битрейта почти в два раза (вместо лока на 192кб),
		// изза чего конвертация проесходит либо без, либо почти без потерь
		// но и вес файла увеличивается тоже два раза

		// бонусом строчка "--postprocessor-args", "ExtractAudio:-af volume=-1.0dB"
		// дополнительно спасает от того самого true peak (если верить gimini).

		cmd = exec.Command(
			"yt-dlp",
			"--newline",
			"--js-runtimes", "node",
			"-f", "ba",
			"-x",
			"--audio-format", "m4a",
			"--audio-quality", "0",
			"--download-sections", downloadSections,
			"--postprocessor-args", "ExtractAudio:-af volume=-1.0dB",
			"--write-thumbnail",           // Скачать обложку в отдельный файл
			"--embed-thumbnail",           // Вшить внутрь m4a
			"--convert-thumbnails", "jpg", // Формат JPG для совместимости
			"-o", outputTemplate,
			url,
		)
	} else {

		// (скачивает в обычном (тоже крутом) качестве)

		cmd = exec.Command(
			"yt-dlp",
			"--newline",
			"--js-runtimes", "node",
			"-f", "ba",
			"-x",
			"--audio-format", "m4a",
			"--audio-quality", "192K",
			"--download-sections", downloadSections,
			"--write-thumbnail",           // Скачать обложку в отдельный файл
			"--embed-thumbnail",           // Вшить внутрь m4a
			"--convert-thumbnails", "jpg", // Формат JPG для совместимости
			"-o", outputTemplate,
			url,
		)
	}

	pipeReader, pipeWriter := io.Pipe()

	// подключаем врайтер к выходам запускаемой проги
	cmd.Stdout = pipeWriter
	cmd.Stderr = pipeWriter

	// запускаем cmd
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

		// обрабатываем только строки с выводом прогресса и выводим в консоль
		if strings.Contains(line, "[download]") && strings.Contains(line, "%") {
			fmt.Print(Yellow + "\ryt-dlp: " + Reset + line + "          ")
		}
	}

	fmt.Print("\n")

	// Если yt-dlp завершился с ошибкой, CloseWithError передаст её в scanner.Err()
	if err := scanner.Err(); err != nil {
		log.Printf("Ошибка при скачивании: %v", err)
		return downloadLog, err
	}

	return downloadLog, nil
}
