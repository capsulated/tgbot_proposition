package scheduler

import (
	"context"
	"gitlab.com/logiq.one/agenda_3dru_bot/telegram"
	"gitlab.com/logiq.one/agenda_3dru_bot/vault"
	"log"
	"time"
)

const runWeekday = 5
const runHour = 10

type Cron struct {
	Vlt        *vault.Postgres
	Bot        *telegram.Bot
	TimeLayout string
}

func New(vlt *vault.Postgres, bot *telegram.Bot) *Cron {
	return &Cron{
		Vlt:        vlt,
		Bot:        bot,
		TimeLayout: "2022-04-16 22:00:00",
	}
}

func (c *Cron) Run() {
	ctx := context.Background()

	nowWeekday := time.Now().Weekday()
	nowHour := time.Now().Hour()
	var diff int

	if nowWeekday < runWeekday {
		diff = runWeekday - int(nowWeekday)
	}
	if nowWeekday == runWeekday && nowHour < runHour {
		diff = 0
	}
	if (nowWeekday == runWeekday && nowHour > runHour) || nowWeekday > runWeekday {
		diff = runWeekday - int(nowWeekday) + 7
	}

	year, month, day := time.Now().Date()
	startTime := time.Date(year, month, day+diff, runHour, 0, 0, 0, time.Now().Location())
	log.Println("SCHEDULER START:", startTime)

	// Delay
	delay := time.Hour * 24 * 7 // 1 week

	for range c.schedule(ctx, startTime, delay) {
		wonInitiative, err := c.Vlt.ListWonInitiatives(startTime.Add(-7 * 24 * time.Hour))
		if err != nil {
			log.Println(err.Error())
			continue
		}

		err = c.Bot.SendWonInitiatives(wonInitiative)
		if err != nil {
			log.Println(err.Error())
		}
	}
}

func (*Cron) schedule(ctx context.Context, startTime time.Time, delay time.Duration) <-chan time.Time {
	// Create the channel which we will return
	stream := make(chan time.Time, 1)

	// Calculating the first start time in the future
	// Need to check if the time is zero (e.g. if time.Time{} was used)
	if !startTime.IsZero() {
		diff := time.Until(startTime)
		if diff < 0 {
			total := diff - delay
			times := total / delay * -1

			startTime = startTime.Add(times * delay)
		}
	}

	// Run this in a goroutine, or our function will block until the first event
	go func() {

		// Run the first event after it gets to the start time
		t := <-time.After(time.Until(startTime))
		stream <- t

		// Open a new ticker
		ticker := time.NewTicker(delay)
		// Make sure to stop the ticker when we're done
		defer ticker.Stop()

		// Listen on both the ticker and the context done channel to know when to stop
		for {
			select {
			case t2 := <-ticker.C:
				stream <- t2
			case <-ctx.Done():
				close(stream)
				return
			}
		}
	}()

	return stream
}
