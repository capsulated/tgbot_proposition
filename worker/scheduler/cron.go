package scheduler

import (
	"context"
	"time"
)

type Cron struct {
	TimeLayout string
}

func New() *Cron {
	return &Cron{TimeLayout: "2022-04-16 22:00:00"}
}

func (c *Cron) Run(work func()) {
	ctx := context.Background()

	startTime, err := time.Parse(c.TimeLayout, "2022-04-16 22:00:00")
	if err != nil {
		panic(err)
	}

	delay := time.Hour * 24 * 7 // 1 week

	for range c.schedule(ctx, startTime, delay) {
		work()
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
