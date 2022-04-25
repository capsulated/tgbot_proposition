package config

import (
	"os"
	"strconv"
)

type App struct {
	Telegram *telegram
	Database *database
	Pin      *pin
}

type database struct {
	ConnStr string
}

type telegram struct {
	ApiToken      string
	MainChannelId int64
	Debug         bool
}

type pin struct {
	Initiator string
	Secretary string
}

func New() *App {
	d := &database{os.Getenv("DB_CONN_STRING")}
	// log.Println(database.ConnStr)

	telegramDebug := os.Getenv("TELEGRAM_DEBUG")
	var debug bool
	if telegramDebug == "true" {
		debug = true
	}

	mainChannelId, _ := strconv.Atoi(os.Getenv("TELEGRAM_INFO_CHANNEL_ID"))
	t := &telegram{
		ApiToken:      os.Getenv("TELEGRAM_API_TOKEN"),
		MainChannelId: int64(mainChannelId),
		Debug:         debug,
	}

	p := &pin{
		Initiator: os.Getenv("PIN_INITIATOR"),
		Secretary: os.Getenv("PIN_SECRETARY"),
	}

	return &App{
		Database: d,
		Telegram: t,
		Pin:      p,
	}
}
