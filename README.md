# Tg Audio Downloader Bot

Тгбот на Go для скачивания аудиодорожек из YouTube, принимает ссылку на видео и отправляет файлом в .m4a. Использует `yt-dlp` для скачивания и `ffmpeg` для конвертации (подробнее ниже)

### Использование
чтобы начать скачиваем готовый бинарник и запускаем (или [собираем свой из исходников](#запуск-и-компиляция-исходников)) При первом запуске бот создаст рядом `config.toml`, заполняем и перезапускаем. После в чате с ботом будет доступен следующий функционал

| Команда / Флаг | Описание |
| :--- | :--- |
| `<ссылка>` | Скачивает аудио по ссылке и отправляет `.m4a` файлом в чат |
| `-l` | Флаг к ссылке. Отправляет текстовый лог работы `yt-dlp` вместе с аудио |
| `/info` | Выводит информацию о боте |
| `*начало-конец` | Скачивает указаный кусок из видео. вместо конкретного времени формата `12:34-12:34` можно указать `0` с начала, `inf` до конца  |

### Качество 
на ютубе аудио хранится преимущественно в `.webm (кодек opus)`, он эффективнее чем необходимый нам `.m4a (кодек AAC)` (первый тг воспринимает как странные головые в отличии от .m4a), при конвертации yt-dlp (ffmpeg) старается сохранить размер файла изза чего детали теряются, сохранить их можно убрав ограничение в `config.toml` значением `song_quality`:
- `song_quality = "default"` - сохраняет размер **( ~192 Kbps, 2 - 4 Mb )**
- `song_quality = "max"` - минимизирует потери **( 400+ Kbps, 6 - 10 Mb )**

Настройки вызова `yt-dlp` находятся в файле [ytdlp.go](./internal/downloader/ytdlp.go)

# Запуск и сборка исходников 

| Зависимость | Описание | Установка (Debian / Ubuntu) |
| :--- | :--- | :--- |
| **Go** (1.21+) | Компилятор и среда выполнения Go | `sudo apt install golang-go` *(или с [официального сайта](https://go.dev/doc/install))* |
| **yt-dlp** | Загрузка медиапотоков с YouTube | `sudo wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -O /usr/local/bin/yt-dlp && sudo chmod a+rx /usr/local/bin/yt-dlp` |
| **ffmpeg** | Конвертация и извлечение чистой аудиодорожки в M4A | `sudo apt install ffmpeg` |
| **Node.js** | Исполнение JS-скриптов YouTube (для флага `--js-runtimes node` и решения n-sig) | `sudo apt install nodejs` |
| **telegram-bot-api** | Библиотека Go для работы с Telegram Bot API | `go get -u github.com/go-telegram-bot-api/telegram-bot-api/v5` |
| **cookies.txt** *(опционально)* | Файл авторизации в корне проекта для обхода блокировок YouTube | Экспорт из браузера через расширение [Get cookies.txt LOCALLY](https://chromewebstore.google.com/detail/get-cookiestxt-locally/cclelndahbckbenkjhflpdbgdldlbecc) |

---

### Сборка и запуск
1. клонируем репозиторий 
```
git clone https://github.com/duckman3310/tgbot.git
```
2. переходим в директорию с main.go
```
cd cmd/bot
```
2. компелируем бинарник
```
go build -o tg-audio-bot 
```
3. запускаем 
```
./tg-audio-bot
```
