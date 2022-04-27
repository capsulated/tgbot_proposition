package config

import (
	"os"
	"strconv"
)

type App struct {
	Telegram  *telegram
	Database  *database
	Pin       *pin
	Scheduler *scheduler
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

type scheduler struct {
	RunWeekday int
	RunHour    int
	Delay      int // minutes
}

func New() *App {
	d := &database{os.Getenv("DB_CONN_STRING")}

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

	runWeekday, err := strconv.Atoi(os.Getenv("SCHEDULER_RUN_WEEKDAY"))
	if err != nil {
		panic(err)
	}

	runHour, err := strconv.Atoi(os.Getenv("SCHEDULER_RUN_HOUR"))
	if err != nil {
		panic(err)
	}

	delay, err := strconv.Atoi(os.Getenv("SCHEDULER_DELAY"))
	if err != nil {
		panic(err)
	}

	s := &scheduler{
		RunWeekday: runWeekday,
		RunHour:    runHour,
		Delay:      delay,
	}

	return &App{
		Database:  d,
		Telegram:  t,
		Pin:       p,
		Scheduler: s,
	}
}
