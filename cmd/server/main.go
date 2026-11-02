package main

import (
	"log"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"genshin-quiz/config"
	"genshin-quiz/internal/enum"
	"genshin-quiz/internal/webserver"
	"genshin-quiz/logger"
)

func main() {
	// Load environment variables based on environment
	envStr := os.Getenv("ENVIRONMENT")
	if envStr == "" {
		envStr = string(enum.DEV) // 默认为开发环境
	}

	env := enum.Environment(envStr)
	var envFile string
	switch env {
	case enum.DEV:
		envFile = ".env.dev"
	case enum.TEST:
		envFile = ".env.test"
	}

	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			log.Printf("Warning: Error loading %s file: %v", envFile, err)
		}
	}

	// Initialize configuration
	app := config.NewApp()

	defer sentry.Flush(2 * time.Second)

	// Set up defer immediately after logger is created
	defer logger.Sync()

	// SERVICE_MODE=auth/quiz/all 用于同一代码库拆分独立进程；
	// 未设置时保持原有单进程行为。
	mode := webserver.Mode(os.Getenv("SERVICE_MODE"))
	if mode == "" {
		mode = webserver.ModeAll
	}

	// Initialize server
	server := webserver.NewServerWithMode(app, mode)
	// Start server in a goroutine
	server.Start()
}
