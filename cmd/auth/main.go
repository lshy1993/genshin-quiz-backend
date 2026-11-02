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

// auth-service：只提供 /auth/* 与 /health。
//
// 与 cmd/server 共用同一个代码库和数据库（第一阶段），但作为独立进程/容器运行，
// 因此登录、注册、OAuth、JWT 签发的重启不再影响 quiz 业务，反之亦然。
func main() {
	envStr := os.Getenv("ENVIRONMENT")
	if envStr == "" {
		envStr = string(enum.DEV)
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

	app := config.NewApp()
	defer sentry.Flush(2 * time.Second)
	defer logger.Sync()

	server := webserver.NewServerWithMode(app, webserver.ModeAuth)
	server.Start()
}
