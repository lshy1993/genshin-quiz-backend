package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"genshin-quiz/config"
	"genshin-quiz/internal/bootstrap"
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

	// 首次部署引导：只有系统里没有任何管理员时才生效。
	// 使用邮箱而不是 user_id，避免换库/重建后 ID 变化。
	if email := os.Getenv("BOOTSTRAP_ADMIN_EMAIL"); email != "" {
		resetPassword, _ := strconv.ParseBool(os.Getenv("BOOTSTRAP_ADMIN_RESET_PASSWORD"))
		if err := bootstrap.EnsureInitialAdmin(
			context.Background(),
			app,
			email,
			os.Getenv("BOOTSTRAP_ADMIN_PASSWORD"),
			os.Getenv("BOOTSTRAP_ADMIN_NICKNAME"),
			resetPassword,
		); err != nil {
			app.Logger.Error("Bootstrap admin failed", zap.Error(err))
		}
	}

	server := webserver.NewServerWithMode(app, webserver.ModeAuth)
	server.Start()
}
