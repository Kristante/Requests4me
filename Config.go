package main

import "time"

type Config struct {
	ChatID     int64
	TickerTime time.Duration
}

func InitConfig() *Config {
	return &Config{
		ChatID:     1062210573,
		TickerTime: 30,
	}
}
