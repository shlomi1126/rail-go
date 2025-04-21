package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"rail-go/internal/bot"
	"rail-go/internal/config"
	"rail-go/internal/logger"
	"rail-go/internal/scheduler"
)

func main() {

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// Initialize logger
	log := logger.New()

	// Initialize bot service
	botService, err := bot.NewService(cfg, log)
	if err != nil {
		log.Fatal("failed to initialize bot service", "error", err)
	}

	// Initialize scheduler
	sched := scheduler.New(log)

	// Create monthly notification task
	if err := sched.ScheduleMonthlyTask(ctx, botService, cfg.Notifications.DefaultChatID); err != nil {
		log.Fatal("failed to schedule monthly task", "error", err)
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("received shutdown signal")
		cancel()
	}()

	// Start the bot service
	if err := botService.Start(ctx); err != nil {
		log.Fatal("bot service failed", "error", err)
	}
}
