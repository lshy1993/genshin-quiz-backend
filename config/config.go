package config

import (
	"context"
	"database/sql"
	"fmt"
	"genshin-quiz/internal/enum"
	"genshin-quiz/logger"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/jwtauth/v5"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/resend/resend-go/v3"
	"go.uber.org/zap"
)

type App struct {
	DB      *sql.DB
	Redis   *redis.Client
	JWTAuth *jwtauth.JWTAuth
	Logger  *zap.Logger
	Storage *azblob.SharedKeyCredential
	Resend  *resend.Client

	Config   AppConfig
	Database DatabaseConfig
	// Worker   WorkerConfig
	Azure  AzureConfig
	Server ServerConfig
}

type AppConfig struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	Environment enum.Environment
	Version     string
	SentryDSN   string
	Domain      string
	ResendKey   string
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type WorkerConfig struct {
	Concurrency int
}

type AzureConfig struct {
	StorageAccount string
	StorageKey     string
	ContainerName  string
}

type ServerConfig struct {
	Host         string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsEnv(key string) enum.Environment {
	val := os.Getenv(key)

	env := enum.Environment(val)
	// 校验获取到的字符串是否属于合法的枚举
	switch env {
	case enum.DEV, enum.TEST, enum.PROD:
		return env
	default:
		return enum.DEV
	}
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue string) time.Duration {
	valueStr := getEnv(key, defaultValue)
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	if value, err := time.ParseDuration(defaultValue); err == nil {
		return value
	}
	return time.Minute * 5
}

func (app *App) initializeDatabase() (*sql.DB, error) {
	// Build connection string from config
	var dsn string
	if app.Config.DatabaseURL != "" {
		dsn = app.Config.DatabaseURL
	} else {
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			app.Database.Host,
			app.Database.Port,
			app.Database.User,
			app.Database.Password,
			app.Database.Name,
		)
	}

	app.Logger.Debug("Connecting to database...")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(app.Database.MaxOpenConns)
	db.SetMaxIdleConns(app.Database.MaxIdleConns)
	db.SetConnMaxLifetime(app.Database.ConnMaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	app.Logger.Info("Database connection established successfully")
	return db, nil
}

func (app *App) initializeSentry() error {
	// 只在非 debug 环境（生产环境）且设置了 Sentry DSN 时才初始化
	// if app.Config.Environment != enum.PROD {
	// 	app.Logger.Info("Sentry disabled in non-production environment")
	// 	return nil
	// }

	if app.Config.SentryDSN == "" {
		app.Logger.Info("Sentry DSN not configured, skipping Sentry initialization")
		return nil
	}

	var tracesSampleRate float64
	if app.Config.Environment == enum.PROD {
		tracesSampleRate = 0.2 // 生产环境采样 20%
	} else {
		tracesSampleRate = 1.0 // 开发环境全量收集/调试
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              app.Config.SentryDSN,
		Environment:      string(app.Config.Environment),
		Debug:            app.Config.Environment == enum.DEV,
		SampleRate:       1.0,              // 在生产环境中可能需要调整采样率
		TracesSampleRate: tracesSampleRate, // 可选，开启 tracing
		EnableTracing:    true,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize Sentry: %w", err)
	}

	app.Logger.Info("Sentry initialized successfully",
		zap.String("environment", string(app.Config.Environment)),
		zap.String("version", app.Config.Version),
	)
	return nil
}

func (app *App) initializeResend() *resend.Client {
	if app.Config.ResendKey == "" {
		app.Logger.Info("Resend key not configured, CANNOT use mail service")
		return nil
	}
	client := resend.NewClient(app.Config.ResendKey)
	app.Logger.Info("Resend mail service initialized successfully")
	return client
}

func (app *App) getFromAddress() string {
	if app.Config.Environment == enum.DEV || app.Config.Domain == "localhost" {
		// 开发环境下使用 Resend 官方测试地址
		return "YourApp Dev <onboarding@resend.dev>"
	}

	// Staging / Production 使用真实域名
	return fmt.Sprintf("YourApp <noreply@%s>", app.Config.Domain)
}

func (app *App) getToAddress(to string) string {
	if app.Config.Environment == enum.DEV || app.Config.Domain == "localhost" {
		return "lshy1993@live.com"
	}
	return to
}

func (app *App) SendEmail(to, subject, htmlBody string) error {
	// 防护拦截：如果邮件服务未正常挂载，记录日志并直接安全退出
	if app.Resend == nil {
		app.Logger.Warn("Email service is not configured/available, skipping email send")
		return nil
	}

	params := &resend.SendEmailRequest{
		From:    app.getFromAddress(),
		To:      []string{app.getToAddress(to)},
		Subject: subject,
		Html:    htmlBody,
	}

	sent, err := app.Resend.Emails.Send(params)
	if err != nil {
		app.Logger.Error("Failed to send email via Resend", zap.Error(err))
		return err
	}

	app.Logger.Info("Email sent successfully",
		zap.String("id", sent.Id),
		zap.String("to", to),
	)
	return nil
}

func NewApp() *App {
	app := &App{
		Config: AppConfig{
			Port: getEnv("PORT", "8080"),
			DatabaseURL: getEnv(
				"DATABASE_URL",
				"postgres://user:password@localhost/genshin_quiz?sslmode=disable",
			),
			JWTSecret:   getEnv("JWT_SECRET", "your-secret-key"),
			Environment: getEnvAsEnv("ENVIRONMENT"),
			Version:     getEnv("VERSION", "dev"),
			SentryDSN:   getEnv("SENTRY_DSN", ""),
			Domain:      getEnv("APP_DOMAIN", "http://localhost:3000"),
			ResendKey:   getEnv("RESEND_KEY", ""),
		},

		Database: DatabaseConfig{
			Host:            getEnv("DATABASE_HOST", "localhost"),
			Port:            getEnv("DATABASE_PORT", "5432"),
			User:            getEnv("DATABASE_USER", "user"),
			Password:        getEnv("DATABASE_PASSWORD", "password"),
			Name:            getEnv("DATABASE_NAME", "genshin_quiz"),
			MaxOpenConns:    getEnvAsInt("DATABASE_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DATABASE_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getEnvAsDuration("DATABASE_CONN_MAX_LIFETIME", "5m"),
		},

		Azure: AzureConfig{
			StorageAccount: getEnv("AZURE_STORAGE_ACCOUNT", ""),
			StorageKey:     getEnv("AZURE_STORAGE_KEY", ""),
			ContainerName:  getEnv("AZURE_CONTAINER_NAME", "genshin-quiz"),
		},

		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", "30s"),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", "30s"),
		},
	}

	if err := logger.Init(string(app.Config.Environment)); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	app.Logger = logger.L

	app.Logger.Info("Current App Config", zap.Any("config", app.Config))

	// 初始化sentry
	err := app.initializeSentry()
	if err != nil {
		app.Logger.Error("Failed to initialize Sentry", zap.Error(err))
		// 不要因为 Sentry 初始化失败而崩溃应用
	}

	// 初始化resend邮件服务
	app.Resend = app.initializeResend()

	// pg数据库
	db, err := app.initializeDatabase()
	if err != nil {
		app.Logger.Fatal("Failed to initialize database:", zap.Error(err))
	}
	app.DB = db

	return app
}
