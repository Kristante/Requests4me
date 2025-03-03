package global

import "time"

type Config struct {
	ChatID          int64
	RequestURL      string
	TickerTime      time.Duration
	Filename        string
	Names           map[string]string
	FilenameNotes   string
	BusinessAccount string
}

func InitConfig() *Config {
	return &Config{
		// Сверху чат айди беседы куда присылать, ниже это личка разработчика
		//ChatID: -1002325494550,
		ChatID:     1062210573,
		RequestURL: "https://api.itsm.mos.ru/v1/requests/",
		TickerTime: 20,
		Filename:   "data/requests_log.txt",

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

		BusinessAccount: "Служебная УЗ",
	}
}
