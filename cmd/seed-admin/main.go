package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"genshin-quiz/config"
	"genshin-quiz/internal/bootstrap"
	"genshin-quiz/internal/enum"
	"genshin-quiz/logger"
)

// 显式管理员 seeder。
//
// 用法：
//
//	go run ./cmd/seed-admin -email you@example.com
//	go run ./cmd/seed-admin -email you@example.com -password 'strong-password' -nickname LuckyLee
//
// 已存在的邮箱只会提权，不会修改密码；不存在的邮箱必须提供 -password。
func main() {
	var (
		email         = flag.String("email", "", "管理员邮箱（必填）")
		password      = flag.String("password", "", "邮箱不存在时用于创建账号的初始密码")
		nickname      = flag.String("nickname", "", "可选昵称")
		resetPassword = flag.Bool("reset-password", false, "邮箱已存在时强制把密码重置为 -password")
	)
	flag.Parse()

	if *email == "" {
		flag.Usage()
		os.Exit(1)
	}

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

	if err := bootstrap.EnsureAdmin(
		context.Background(), app, *email, *password, *nickname, *resetPassword,
	); err != nil {
		app.Logger.Error("Seed admin failed", zap.Error(err))
		os.Exit(1)
	}
	app.Logger.Info("Admin seed completed", zap.String("email", *email))
}
