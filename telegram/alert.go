package telegram

import (
	tele "gopkg.in/telebot.v4"
	"log"
)

func SendAlertForChat(bot *tele.Bot, chatID int64, message string) {
	// Отправляем сообщение в чат с ошибками
	_, err := bot.Send(&tele.Chat{ID: chatID}, message)
	if err != nil {
		log.Printf("Не удалось отправить сообщение: %v", err)
	}
}
