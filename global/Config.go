package global

import "time"

type Config struct {
	RobotMainErrorMessage string
	ChatID                int64
	ErrorChatID           int64
	RequestURL            string
	Filename              string
	Names                 map[string]string
	Messages              map[string]string
	FilenameNotes         string
	BusinessAccount       string
}

func InitConfig() *Config {
	return &Config{
		RobotMainErrorMessage: "❌❌❌ FAULTED!\n\n" + time.Now().Format("2006-01-02 15:04:05") + "\n\nПроизошла ошибка в работе робота\n\nProcessName: 4meRequestsBot\n\nExceptionMessage:\n",
		// Сверху чат айди беседы куда присылать, ниже это личка разработчика
		//ChatID:                -1002325494550,
		ChatID:      1062210573,
		ErrorChatID: -4669217347,
		RequestURL:  "https://api.itsm.mos.ru/v1/requests/",
		Filename:    "data/requests_log.txt",

		FilenameNotes: "data/requests_notes.txt",
		Names: map[string]string{
			"Иванов Дмитрий Николаевич":         "@kristante",
			"Булавина Василина Васильевна":      "@vslnb",
			"Дюжов Артём Витальевич":            "Заявка Артёма",
			"Епифановский Михаил Александрович": "Заявка Миши",
			"Ершов Александр Павлович":          "Заявка Саши",
			"Любимов Георгий Владимирович":      "Заявка Гоши",
			"Потриваев Никита Андреевич":        "Заявка Никиты",
			"Хасаншин Рустам Альбертович":       "Заявка Рустама",
			"Хусниярова Алия Раисовна":          "Заявка Алии",
			"Cлужебная УЗ":                      "",
			"Автоматизация":                     ""},
		Messages: map[string]string{
			"errorCreateBot": "Бот 4meRequests не доступен",
			"OsDownload":     "Не удалось прочитать .env файл",
		},
		BusinessAccount: "Служебная УЗ",
	}
}
