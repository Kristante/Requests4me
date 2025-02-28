package handlers4me

import (
	"4meRequests/global"
	"4meRequests/telegram"
	"encoding/json"
	"fmt"
	tele "gopkg.in/telebot.v4"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"
)

type ProcessedComments struct {
	RequestID  int
	CommentIDs map[int]struct{}
}

type Person struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Account struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"account"`
	SourceID string `json:"sourceID,omitempty"`
	NodeID   string `json:"nodeID"`
}

type Attachment struct {
	CreatedAt string `json:"created_at"`
	ID        int    `json:"id"`
	Inline    bool   `json:"inline"`
	Key       string `json:"key"`
	Name      string `json:"name"`
	Size      int    `json:"size"`
	URI       string `json:"uri"`
	NodeID    string `json:"nodeID"`
	Note      struct {
		ID     int    `json:"id"`
		NodeID string `json:"nodeID"`
	} `json:"note"`
}

type Note struct {
	ID          int          `json:"id"`
	Person      Person       `json:"person"`
	CreatedAt   string       `json:"created_at"`
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
	Medium      string       `json:"medium"`
	Internal    bool         `json:"internal"`
	Account     struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"account"`
	NodeID string `json:"nodeID"`
}

func GetNotesFromRequest(reqID int, apiToken string, bot *tele.Bot, memberName string, businessAcc bool) {
	apiURL := "https://api.itsm.mos.ru/v1/requests/" + strconv.Itoa(reqID) + "/notes"
	client := http.Client{}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.Printf("Ошибка создания запроса: %v", err)
		return
	}

	// Устанавливаем заголовки
	req.Header.Add("Authorization", "Bearer "+apiToken)
	req.Header.Add("x-4me-account", "rpa")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		return
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка чтения тела ответа: %v", err)
		return
	}

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		log.Printf("Ошибка: %s. Тело ответа: %s", resp.Status, string(body))
		return
	}

	// Распарсить JSON
	var notes []Note
	if err := json.Unmarshal(body, &notes); err != nil {
		log.Printf("Ошибка парсинга JSON: %v", err)
		return
	}

	if businessAcc == true {

		sort.Slice(notes, func(i, j int) bool {
			t1, _ := time.Parse(time.RFC3339, notes[i].CreatedAt)
			t2, _ := time.Parse(time.RFC3339, notes[j].CreatedAt)
			return t2.After(t1) // Сортировка по возрастанию
		})

		config := global.InitConfig()

		if len(notes) > 0 {
			latestNote := notes[0]

			fmt.Printf("Кто создал: %s\nТекст: %s\n", latestNote.Person.Name, latestNote.Text)
			message := fmt.Sprintf("%s\nПоявилась новая автоматическая заявка от пользователя %s под номером %d:\nТекст комментария: %s\nОзнакомиться подробнее: https://rpa.itsm.mos.ru/requests/%d", latestNote.Person.Name, reqID, latestNote.Text, reqID)
			telegram.SendMessageForChat(bot, config.ChatID, message)
		}

	} else {

		// Сортировка по дате создания
		sort.Slice(notes, func(i, j int) bool {
			t1, _ := time.Parse(time.RFC3339, notes[i].CreatedAt)
			t2, _ := time.Parse(time.RFC3339, notes[j].CreatedAt)
			return t1.After(t2) // Сортировка по убыванию
		})

		config := global.InitConfig()
		excludedNames := config.Names
		// Список имен для исключения
		// Самая последняя заметка
		if len(notes) > 0 {
			latestNote := notes[0]

			// Проверка на отсутствие имени в списке исключений
			nameAllowed := true

			// Проверяем, существует ли имя в ключах мапы
			if _, exists := excludedNames[latestNote.Person.Name]; exists {
				nameAllowed = false
			}

			if nameAllowed {
				link := excludedNames[memberName]
				fmt.Printf("Кто создал: %s\nТекст: %s\n", latestNote.Person.Name, latestNote.Text)
				message := fmt.Sprintf("%s\nПоявился новый комментарий от пользователя %s у заявки под номером %d:\nТекст комментария: %s\nОзнакомиться подробнее: https://rpa.itsm.mos.ru/requests/%d", link, latestNote.Person.Name, reqID, latestNote.Text, reqID)
				telegram.SendMessageForChat(bot, config.ChatID, message)
			} else {
				fmt.Println("Имя создателя входит в список исключений, выводить информацию не нужно.")
			}
		} else {
			fmt.Println("Заметок нет.")
		}
	}

}
