package bot

import (
	"fmt"
	"log"
	"sync"
)

type ChatLogRing struct {
	mu     sync.Mutex
	chatID int64
	pos    int
	size   int
	logs   []string
}

type LogStore struct {
	mu   sync.Mutex
	logs []*ChatLogRing
}

func NewChatLogRing(ChatID int64, Size int) *ChatLogRing {
	return &ChatLogRing{
		chatID: ChatID,
		size:   Size,
		logs:   make([]string, Size),
	}
}

func (ring *ChatLogRing) Add(log string) {

	ring.mu.Lock()
	defer ring.mu.Unlock()

	ring.logs[ring.pos] = log
	ring.pos++

	if ring.pos >= ring.size {
		ring.pos = 0
	}
}

func (ring *ChatLogRing) Get() []string {

	ring.mu.Lock()
	defer ring.mu.Unlock()

	pos := ring.pos + 1
	var result []string

	for i := 0; i < len(ring.logs); i++ {

		if ring.logs[pos] != "" {
			result = append(result, ring.logs[pos])
		}

		pos++

		if pos > len(ring.logs)-1 {
			pos = 0
		}
	}

	return result
}
func (ls *LogStore) Write(ChatID int64, line string) {

	ls.mu.Lock()
	defer ls.mu.Unlock()

	for i := 0; i < len(ls.logs); i++ {
		if ls.logs[i].chatID == ChatID {
			ls.logs[i].Add(line)
			return
		}
	}

	ls.logs = append(ls.logs, NewChatLogRing(ChatID, 15))
	ls.logs[len(ls.logs)-1].Add(line)
}

func (ls *LogStore) Read(ChatID int64) []string {

	ls.mu.Lock()
	defer ls.mu.Unlock()

	for i := 0; i < len(ls.logs); i++ {
		if ls.logs[i].chatID == ChatID {
			result := ls.logs[i].Get()
			return result
		}
	}

	return []string{}
}

func logBot(description string, ChatID ...int64) {
	logtext := fmt.Sprintf("%v Info: %s", ChatID, description)
	log.Print(Green + logtext + Reset)
	if len(ChatID) > 0 {
		UserLogs.Write(ChatID[0], logtext)
	}
}
func logErr(description string, err error, ChatID ...int64) {
	logtext := fmt.Sprintf("%v Error: %s: %v", ChatID, description, err)
	log.Print(Red + logtext + Reset)
	if len(ChatID) > 0 {
		UserLogs.Write(ChatID[0], logtext)
	}
}

func logWarn(text string, ChatID ...int64) {
	logtext := fmt.Sprintf("%v Warning: %s", ChatID, text)
	log.Print(Yellow + logtext + Reset)
	if len(ChatID) > 0 {
		UserLogs.Write(ChatID[0], logtext)
	}
}
