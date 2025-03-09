package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"time"
)

func nextMonthlyRun(now time.Time) time.Time {
	year, month, _ := now.Date()
	loc := now.Location()

	// Calculate the last day of the current month.
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, loc)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	durationUntilEnd := lastOfMonth.Sub(now)
	var scheduledTime time.Time
	if durationUntilEnd.Hours() < 120.0 {
		daysLeft := int(durationUntilEnd.Hours() / 24)
		lastOfMonth = firstOfMonth.AddDate(0, 1, -daysLeft)
		scheduledTime = time.Date(year, month, lastOfMonth.Day(), 0, 0, 0, 0, loc)

	} else {
		scheduledTime = time.Date(year, month, lastOfMonth.Day()-5, 0, 0, 0, 0, loc)
	}

	return scheduledTime
}

// scheduler calculates the delay until the next scheduled time (last day of the month)
// and then waits on a timer. When the timer fires, it sends the execution time
// over the run channel.
// A stop channel is used to shut it down gracefully.
func scheduler(run chan<- time.Time, stop <-chan struct{}) {
	for {
		now := time.Date(2025, time.March, 28, 0, 0, 0, 0, time.UTC)
		nextRunTime := nextMonthlyRun(now)
		duration := time.Until(nextRunTime)
		fmt.Printf("Scheduler: next run scheduled at %v (in %v)\n", nextRunTime, duration)

		// Create a timer that will expire when it's time to run.
		timer := time.NewTimer(duration)
		select {
		case t := <-timer.C:
			// When the timer fires, send the time over the run channel.
			run <- t
		case <-stop:
			timer.Stop()
			close(run)
			return
		}
	}
}

// monthlyTask waits for the run channel to signal when it’s time to execute,
// then performs the work (here, simply printing messages).
func monthlyTask(b *Bot, run <-chan time.Time, done chan<- struct{}) {
	for t := range run {
		msg := tgbotapi.NewMessageToChannel("shlomi1126", t.Local().Format("Mon Jan 2 15:04:05"))
		b.send(msg)
		done <- struct{}{}
	}
}
