package telegram

import (
	"github.com/joho/godotenv"
	tele "gopkg.in/telebot.v4"
	"log"
	"os"
	"time"
)

func CreateBot() (*tele.Bot, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки .env файла: %v", err)
		return nil, err
	}

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN не найден в файле .env")
		return nil, err
	}

	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}
	bot, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return bot, nil
}

func SendMessageForChat(bot *tele.Bot, chatID int64, message string) {
	// Отправляем сообщение в указанный чат
	_, err := bot.Send(&tele.Chat{ID: chatID}, message)
	if err != nil {
		log.Printf("Не удалось отправить сообщение: %v", err)
	}
}
