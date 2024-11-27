package main

import (
	"4meRequests/telegram"

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
	ID      int    `json:"id"`
	Subject string `json:"subject"`
	Member  struct {
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
}

// Структуры для парсинга JSON
type CustomField struct {
	ID    string          `json:"id"`
	Value json.RawMessage `json:"value"`
}

type CreatedBy struct {
	Name string `json:"name"`
}

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

func getInfoForRequest(reqID int, apiToken string, bot *tele.Bot) {
	apiURL := "https://api.itsm.mos.ru/v1/requests/" + strconv.Itoa(reqID)
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
		log.Fatalf("Ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Ошибка чтения тела ответа: %v", err)
	}

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Ошибка: %s. Тело ответа: %s", resp.Status, string(body))
	}

	// Распарсить JSON
	var requests Request
	if err := json.Unmarshal(body, &requests); err != nil {
		log.Fatalf("Ошибка парсинга JSON: %v", err)
	}

	var description string
	for _, field := range requests.CustomFields {
		if field.ID == "description" {
			if field.Value != nil {
				var value string
				if err := json.Unmarshal(field.Value, &value); err != nil {
					log.Printf("Ошибка парсинга значения поля 'description': %v", err)
				} else {
					description = value
				}

			}
		}

	}

	description = cleanString(description)

	config := InitConfig()

	NameOfCreator := fmt.Sprintf("%v", requests.CreatedBy)
	NameOfCreator = strings.Trim(NameOfCreator, "{}")
	IDRequest := requests.ID

	message := fmt.Sprintf("Заявка: %d - %s\n /n %s", IDRequest, NameOfCreator, description)
	telegram.SendMessageForChat(bot, config.ChatID, message)

}

// Функция для получения заявок
func getAllRequests(apiToken string, bot *tele.Bot) {
	if apiToken == "" {
		log.Fatal("Токен не установлен. Проверьте переменную окружения.")
	}

	apiURL := "https://api.itsm.mos.ru/v1/requests"
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
		log.Fatalf("Ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Ошибка чтения тела ответа: %v", err)
	}

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Ошибка: %s. Тело ответа: %s", resp.Status, string(body))
	}

	// Распарсить JSON
	var requests []Requests
	if err := json.Unmarshal(body, &requests); err != nil {
		log.Fatalf("Ошибка парсинга JSON: %v", err)
	}

	// Выводим все заявки
	for _, req := range requests {
		if req.Member.Name == "" {
			fmt.Printf("Заявка: %d - %s\n", req.ID, req.Subject)
			getInfoForRequest(req.ID, apiToken, bot)

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

	config := InitConfig()

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
