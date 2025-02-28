package main

import (
	"4meRequests/global"
	"4meRequests/handlers4me"
	"4meRequests/telegram"
	"4meRequests/txtfile"

	tele "gopkg.in/telebot.v4"

	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Структура для всех заявок

type Requests struct {
	ID        int       `json:"id"`
	Subject   string    `json:"subject"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
	Member    struct {
		Name string `json:"name"`
	} `json:"member"`
}

type Request struct {
	ID      int    `json:"id"`
	Subject string `json:"subject"`
	Member  struct {
		Name string `json:"name"`
	} `json:"member"`
	CustomFields []CustomField `json:"custom_fields"`
	CreatedBy    CreatedBy     `json:"created_by"`
	ReopenCount  int           `json:"reopen_count"`
}

// Структуры для парсинга JSON
type CustomField struct {
	ID    string          `json:"id"`
	Value json.RawMessage `json:"value"`
}

type CreatedBy struct {
	Name string `json:"name"`
}

// Функция чистим от всяких символов строку
func cleanString(input string) string {
	// Регулярное выражение для замены всех последовательностей вида "/<буква>" на пробел
	re := regexp.MustCompile(`/[a-zA-Z]`) // Ищет "/a", "/b", "/n" и так далее

	// Заменяем все найденные последовательности на пробел
	input = re.ReplaceAllString(input, " ")

	// Регулярное выражение для удаления символов новой строки, табуляции и других управляющих символов.
	// Оставляем только пробелы, буквы и цифры.
	re2 := regexp.MustCompile(`[\r\n\t]+`) // Убираем только \r, \n, \t

	// Заменяем все найденные символы на пробел
	cleaned := re2.ReplaceAllString(input, " ")

	// Убираем лишние пробелы, если они есть (например, несколько пробелов подряд)
	cleaned = strings.Join(strings.Fields(cleaned), " ")

	// Возвращаем очищенную строку
	return cleaned
}

// Получить информацию о конкретной
func getInfoForRequest(reqID int, apiToken string, bot *tele.Bot) {
	apiURL := "https://api.itsm.mos.ru/v1/requests/" + strconv.Itoa(reqID)
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
	var requests Request
	if err := json.Unmarshal(body, &requests); err != nil {
		log.Printf("Ошибка парсинга JSON: %v", err)
		return
	}

	var description string
	for _, field := range requests.CustomFields {
		if field.ID == "system_id" {
			if field.Value != nil {
				var value string
				if err := json.Unmarshal(field.Value, &value); err != nil {
					log.Printf("Ошибка парсинга значения поля 'description': %v", err)
				} else {

					if value == "" {
						fmt.Printf("ОИВ не указан" + description)
						description = "Робот не указан"
					} else {
						fmt.Printf("ОИВ указан" + description)
						description = value
					}

				}

			}

		}

	}

	description = cleanString(description)

	config := global.InitConfig()

	NameOfCreator := requests.CreatedBy.Name

	if NameOfCreator == config.BusinessAccount {
		handlers4me.GetNotesFromRequest(requests.ID, apiToken, bot, requests.Member.Name, true)
	} else {
		IDRequest := requests.ID

		message := fmt.Sprintf("%s\nПоявилась новая заявка под номером %d от пользователя %s\nОзнакомиться подробнее можно по ссылке: https://rpa.itsm.mos.ru/requests/%d", description, IDRequest, NameOfCreator, IDRequest)
		telegram.SendMessageForChat(bot, config.ChatID, message)
	}
}

// Функция для получения заявок
func getAllRequests(apiToken string, bot *tele.Bot) {
	if apiToken == "" {
		log.Fatal("Токен не установлен. Проверьте переменную окружения.")
	}

	// Подгруждаем с конфига название файла
	config := global.InitConfig()
	filename := config.Filename
	filenameNotes := config.FilenameNotes

	apiURL := "https://api.itsm.mos.ru/v1/requests/open"
	client := http.Client{}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.Fatalf("Ошибка создания запроса: %v", err)
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
	var requests []Requests
	if err := json.Unmarshal(body, &requests); err != nil {
		log.Printf("Ошибка парсинга JSON: %v", err)
		return
	}

	// Выводим все заявки
	for _, req := range requests {
		if req.Member.Name == "" {
			exists, err := txtfile.ContainsRequest(req.ID, filename, "")
			if err != nil {
				log.Printf("Ошибка чтения файла: %v", err)
				continue
			}

			if !exists {
				fmt.Printf("Появилась новая заявка: %d - %s\n /n Ознакомиться подробнее можно по ссылке: https://rpa.itsm.mos.ru/requests/%d", req.ID, req.Subject, req.ID)
				getInfoForRequest(req.ID, apiToken, bot)

				err := txtfile.AddRequestToFile(req.ID, filename, "")
				if err != nil {
					log.Printf("Ошибка записи в файл: %v", err)
				}
			}

		}

		if req.Status == "assigned" && req.Member.Name != "" {
			UpdatedTimeString := req.UpdatedAt.Format("2006-01-02T15:04:05-07:00")
			exists, err := txtfile.ContainsRequest(req.ID, filenameNotes, UpdatedTimeString)
			if err != nil {
				log.Printf("Ошибка чтения файла: %v", err)
				continue
			}
			if !exists {
				handlers4me.GetNotesFromRequest(req.ID, apiToken, bot, req.Member.Name, false)
				err := txtfile.AddRequestToFile(req.ID, filenameNotes, UpdatedTimeString)
				if err != nil {
					log.Printf("Ошибка записи в файл: %v", err)
				}

			}

		}

	}
}

func main() {

	bot, err := telegram.CreateBot()
	if err != nil {
		return
	}

	go func() {
		bot.Start()
	}()

	err = godotenv.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки .env файла: %v", err)
	}

	apiToken := os.Getenv("TOKEN_4ME")
	getAllRequests(apiToken, bot)

	config := global.InitConfig()

	ticker := time.NewTicker(config.TickerTime * time.Minute)
	defer ticker.Stop()

	fmt.Println("Запуск планировщика...")
	go func() {
		for {
			select {
			case <-ticker.C:
				// Выполняем функцию каждые 30 минут
				fmt.Println("Таймер сработал, выполняется получение заявок...")
				getAllRequests(apiToken, bot)
			}
		}
	}()

	select {}
}
