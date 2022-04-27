package scheduler

import (
	"context"
	"gitlab.com/logiq.one/agenda_3dru_bot/config"
	"gitlab.com/logiq.one/agenda_3dru_bot/telegram"
	"gitlab.com/logiq.one/agenda_3dru_bot/vault"
	"log"
	"time"
)

type Cron struct {
	Vlt        *vault.Postgres
	Bot        *telegram.Bot
	TimeLayout string
	RunWeekday int
	RunHour    int
	RunMinute  int
	Delay      int
}

func New(cfg *config.App, vlt *vault.Postgres, bot *telegram.Bot) *Cron {
	return &Cron{
		Vlt:        vlt,
		Bot:        bot,
		TimeLayout: "2022-04-16 22:00:00",
		RunWeekday: cfg.Scheduler.RunWeekday,
		RunHour:    cfg.Scheduler.RunHour,
		RunMinute:  cfg.Scheduler.RunMinute,
		Delay:      cfg.Scheduler.Delay,
	}
}

func (c *Cron) Run() {
	ctx := context.Background()

	nowWeekday := time.Now().Weekday()
	nowHour := time.Now().Hour()
	nowMinute := time.Now().Minute()

	var diff int
	if int(nowWeekday) < c.RunWeekday {
		diff = c.RunWeekday - int(nowWeekday)
	}
	if (int(nowWeekday) == c.RunWeekday && nowHour < c.RunHour) || (int(nowWeekday) == c.RunWeekday && nowHour == c.RunHour && nowMinute < c.RunMinute) {
		diff = 0
	}
	if (int(nowWeekday) == c.RunWeekday && nowHour > c.RunHour) || (int(nowWeekday) == c.RunWeekday && nowHour == c.RunHour && nowMinute > c.RunMinute) || int(nowWeekday) > c.RunWeekday {
		diff = c.RunWeekday - int(nowWeekday) + 7
	}

	year, month, day := time.Now().Date()
	startTime := time.Date(year, month, day+diff, c.RunHour, c.RunMinute, 0, 0, time.Now().Location())
	log.Println("SCHEDULER START TIME: ", startTime)

	// Delay
	delay := time.Minute * time.Duration(c.Delay)
	log.Println("SCHEDULER DELAY: ", delay)

	for range c.schedule(ctx, startTime, delay) {
		log.Println("SCHEDULER TICK: ", time.Now())
		log.Println("CREATED_AT > ", startTime.Add(-1*delay))

		wonInitiatives, err := c.Vlt.ListWonInitiatives(startTime.Add(-1 * delay))
		if err != nil {
			log.Println(err.Error())
			continue
		}

		log.Println("WON INITIATIVES TICK: ", time.Now())
		err = c.Bot.SendWonInitiatives(wonInitiatives)
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
