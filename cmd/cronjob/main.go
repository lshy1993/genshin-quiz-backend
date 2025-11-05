package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"genshin-quiz/config"
	"genshin-quiz/internal/cronjob"

	"github.com/robfig/cron/v3"
)

func main() {
	app := config.NewApp()
	cronJob := cronjob.NewCronjob(app)

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <command>")
		fmt.Println("Available commands:")
		fmt.Println("  cronjob:daily-stats    - Run daily statistics recalibration at 12:00")
		fmt.Println("  cronjob:every-five-minutes - Run every 5 minutes (for testing)")
		fmt.Println("  cronjob:run-once       - Run once immediately")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "cronjob:every-five-minutes":
		runEveryFiveMinutes(cronJob)
	case "cronjob:run-once":
		runOnce(cronJob)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

// runEveryFiveMinutes 每5分钟执行一次（测试用）
func runEveryFiveMinutes(cronJob *cronjob.Cronjob) {
	c := cron.New()

	// 每5分钟执行
	_, err := c.AddFunc("*/5 * * * *", func() {
		fmt.Println("Starting 5-minute statistics recalibration...")

		if err := cronJob.RecalibrateUserStats(); err != nil {
			fmt.Printf("Failed to recalibrate user stats: %v\n", err)
		}

		if err := cronJob.RecalibrateQuestionStats(); err != nil {
			fmt.Printf("Failed to recalibrate question stats: %v\n", err)
		}

		fmt.Println("5-minute statistics recalibration completed.")
	})

	if err != nil {
		fmt.Printf("Failed to schedule 5-minute job: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("5-minute cron job started. Will run every 5 minutes.")
	startCronAndWait(c)
}

// runOnce 立即执行一次（测试用）
func runOnce(cronJob *cronjob.Cronjob) {
	fmt.Println("Running statistics recalibration once...")

	if err := cronJob.RecalibrateUserStats(); err != nil {
		fmt.Printf("Failed to recalibrate user stats: %v\n", err)
		os.Exit(1)
	}

	if err := cronJob.RecalibrateQuestionStats(); err != nil {
		fmt.Printf("Failed to recalibrate question stats: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Statistics recalibration completed.")
}

// startCronAndWait 启动 cron 并等待信号
func startCronAndWait(c *cron.Cron) {
	c.Start()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Cron job is running. Press Ctrl+C to stop.")
	<-quit

	fmt.Println("Shutting down cron job...")
	c.Stop()
}
