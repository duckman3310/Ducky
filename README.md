# Ducky

Тгбот на Go для скачивания аудио, принимает ссылку на видео и отправляет файлом .m4a в чат. Сделан впервую очередь под личное использование, но и сбольшим количеством людей проблем не будет

### Использование

| Команда / Флаг | Описание |
| :--- | :--- |
| `ссылка` | Скачивает аудио по ссылке и отправляет `.m4a` файлом в чат |
| `/info` | Выводит информацию о боте |
| `/logs` | Отправляет логи работы в текущем чате|
| `*начало-конец` | Скачивает указаный кусок из видео. формат `12:34-12:34` или `0` с начала, `inf` до конца  |

### Качество 
в чате телеграмма родной .webm (кодек opus) ютуба работает криво, поэтому все отправляется в .m4a (кодек AAC), при конвертации изза разности в эффективности сжатия в `config.toml` можно выбрать сохранять: 
- размер файла `song_quality = "default"`
- качество аудио `song_quality = "max"`  

(За скачивание отвечает [yt-dlp](https://github.com/yt-dlp/yt-dlp), настройки его вызова находятся в файле [ytdlp.go](./internal/downloader/ytdlp.go))

### Запуск
Можно скачать релиз, или собрать самому, одной из комманд ниже. При первом запуске бот создаст рядом `config.toml`, его нужно заполнить и перезапустить

| ОС | Комманда для сборки |
| :--- | :--- |
| linux x86-64 | `GOOS=linux   GOARCH=amd64 go build -o build/ducky-linux-amd64 ./cmd/bot` | 
| linux arm64 | `GOOS=linux   GOARCH=arm64 go build -o build/ducky-linux-arm64 ./cmd/bot` |
| windows x86-64 | `GOOS=windows GOARCH=amd64 go build -o build/ducky-windows-amd64.exe ./cmd/bot` |
| macos x86-64 | `GOOS=darwin  GOARCH=amd64 go build -o build/ducky-darwin-amd64 ./cmd/bot`| 
| macos arm64 | `GOOS=darwin GOARCH=arm64 go build -o build/ducky-darwin-arm64  ./cmd/bot`|
