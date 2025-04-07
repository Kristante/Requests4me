package main

import (
	"4meRequests/global"
	"4meRequests/handlers4me"
	"4meRequests/telegram"
	"4meRequests/txtfile"
	"errors"
	"github.com/joho/godotenv"
	"os"

	tele "gopkg.in/telebot.v4"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
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

// Структура для конкретной заявки
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

func main() {

	config := global.InitConfig()

	bot, err := telegram.CreateBot("BOT_ALERT_TOKEN")
	if err != nil {
		telegram.SendAlertForChat(bot, config.ErrorChatID, config.RobotMainErrorMessage+config.Messages["errorCreateBot"])
	}

	go func() {
		bot.Start()
	}()

	err = godotenv.Load()
	if err != nil {
		telegram.SendAlertForChat(bot, config.ErrorChatID, config.RobotMainErrorMessage+config.Messages["OsDownload"])
		log.Fatalf("Ошибка загрузки .env файла: %v", err)
	}

	apiToken := os.Getenv("TOKEN_4ME")

	err = getAllRequests(apiToken, bot, config)
	if err != nil {
		telegram.SendAlertForChat(bot, config.ErrorChatID, config.RobotMainErrorMessage+fmt.Sprintf("\nПроизошла ошибка %v", err))
	}
}

// Получить информацию о конкретной заявке
func getInfoForRequest(reqID int, apiToken string, bot *tele.Bot) error {

	config := global.InitConfig()
	apiURL := config.RequestURL + strconv.Itoa(reqID)
	client := http.Client{}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.Printf("Ошибка создания запроса: %v", err)
		return errors.New(fmt.Sprintf("Ошибка создания запроса: %v", err))
	}

	// Устанавливаем заголовки
	req.Header.Add("Authorization", "Bearer "+apiToken)
	req.Header.Add("x-4me-account", "rpa")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		return errors.New(fmt.Sprintf("\nОшибка выполнения запроса: %v", err))
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка чтения тела ответа: %v", err)
		return errors.New(fmt.Sprintf("\nОшибка чтения тела ответа: %v", err))
	}

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		log.Printf("Ошибка: %s. Тело ответа: %s", resp.Status, string(body))
		return errors.New(fmt.Sprintf("\nОшибка: %s. Тело ответа: %s", resp.Status, string(body)))
	}

	// Распарсить JSON
	var requests Request
	if err := json.Unmarshal(body, &requests); err != nil {
		log.Printf("Ошибка парсинга JSON: %v", err)
		return errors.New(fmt.Sprintf("\nОшибка парсинга JSON: %v", err))
	}

	NameOfCreator := requests.CreatedBy.Name

	if NameOfCreator == config.BusinessAccount {
		handlers4me.GetNotesFromRequest(requests.ID, apiToken, bot, requests.Member.Name, true, config)
	} else {
		IDRequest := requests.ID

		message := fmt.Sprintf("Появилась новая заявка под номером %d от пользователя %s\nОзнакомиться подробнее можно по ссылке: https://rpa.itsm.mos.ru/requests/%d", IDRequest, NameOfCreator, IDRequest)
		telegram.SendMessageForChat(bot, config.ChatID, message)
	}
	return nil
}

// Функция для получения заявок
func getAllRequests(apiToken string, bot *tele.Bot, config *global.Config) error {

	if apiToken == "" {
		return errors.New("\nне установлен токен 4me")
	}

	// Подгруждаем с конфига название файла

	filename := config.Filename
	filenameNotes := config.FilenameNotes

	apiURL := config.RequestURL + "open"
	client := http.Client{}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return errors.New(fmt.Sprintf("\nОшибка создания запроса: %v", err))
	}

	// Устанавливаем заголовки
	req.Header.Add("Authorization", "Bearer "+apiToken)
	req.Header.Add("x-4me-account", "rpa")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		return errors.New(fmt.Sprintf("\nОшибка выполнения запроса: %v", err))
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка чтения тела ответа: %v", err)
		return errors.New(fmt.Sprintf("\nОшибка чтения тела ответа: %v", err))
	}

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		log.Printf("Ошибка: %s. Тело ответа: %s", resp.Status, string(body))
		return errors.New(fmt.Sprintf("\nОшибка: %s. Тело ответа: %s", resp.Status, string(body)))
	}

	// Распарсить JSON
	var requests []Requests
	if err := json.Unmarshal(body, &requests); err != nil {
		log.Printf("Ошибка парсинга JSON: %v", err)
		return errors.New(fmt.Sprintf("\nОшибка парсинга JSON: %v", err))
	}

	// Выводим все заявки
	for _, req := range requests {
		if req.Member.Name == "" {
			exists, err := txtfile.ContainsRequest(req.ID, filename, "")
			if err != nil {
				log.Printf("\nОшибка чтения файла: %v", err)
				continue
			}

			if !exists {
				fmt.Printf("Появилась новая заявка: %d - %s\n /n Ознакомиться подробнее можно по ссылке: https://rpa.itsm.mos.ru/requests/%d", req.ID, req.Subject, req.ID)
				err = getInfoForRequest(req.ID, apiToken, bot)
				if err != nil {
					return errors.New(fmt.Sprintf("\nНе удалось получить данные о заявке. %v", err))
				}
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
				handlers4me.GetNotesFromRequest(req.ID, apiToken, bot, req.Member.Name, false, config)
				err := txtfile.AddRequestToFile(req.ID, filenameNotes, UpdatedTimeString)
				if err != nil {
					log.Printf("Ошибка записи в файл: %v", err)
					return errors.New(fmt.Sprintf("\nОшибка записи в файл: %v", err))
				}

			}

		}

	}
	return nil
}
