package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/jwtauth/v5"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type App struct {
	DB      *sql.DB
	Redis   *redis.Client
	JWTAuth *jwtauth.JWTAuth
	Logger  *zap.Logger
	Storage *azblob.SharedKeyCredential

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
	Environment string
	Version     string
	SentryDSN   string
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

func (app *App) initializeLogger() (*zap.Logger, error) {
	var config zap.Config

	if app.Config.Environment == "production" {
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		// 生产环境不使用彩色输出
		config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		// 开发环境使用彩色输出
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// 设置更友好的时间格式和颜色编码
	if app.Config.Environment != "production" {
		config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.000")
		// 使用短路径显示 caller 信息
		config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	logger, _ := config.Build()
	// Set up defer immediately after logger is created
	defer func() {
		err := logger.Sync()
		if err != nil &&
			!errors.Is(err, syscall.EINVAL) && // invalid argument
			!errors.Is(err, syscall.EBADF) && // bad file descriptor
			!errors.Is(err, syscall.ENOTTY) {
			panic(err.Error())
		}
	}()

	return logger, nil
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
	// 只在设置了 Sentry DSN 时才初始化
	if app.Config.SentryDSN == "" {
		app.Logger.Info("Sentry DSN not configured, skipping Sentry initialization")
		return nil
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              app.Config.SentryDSN,
		Environment:      app.Config.Environment,
		Debug:            app.Config.Environment == "development",
		SampleRate:       1.0, // 在生产环境中可能需要调整采样率
		TracesSampleRate: 1.0, // 可选，开启 tracing
		EnableTracing:    app.Config.Environment == "production",
	})
	if err != nil {
		return fmt.Errorf("failed to initialize Sentry: %w", err)
	}

	app.Logger.Info("Sentry initialized successfully",
		zap.String("environment", app.Config.Environment),
		zap.String("version", app.Config.Version),
	)
	return nil
}

func NewApp() *App {
	app := &App{
		Config: AppConfig{
			Port:        getEnv("PORT", "8080"),
			DatabaseURL: getEnv("DATABASE_URL", "postgres://user:password@localhost/genshin_quiz?sslmode=disable"),
			JWTSecret:   getEnv("JWT_SECRET", "your-secret-key"),
			Environment: getEnv("ENVIRONMENT", "development"),
			Version:     getEnv("VERSION", "dev"),
			SentryDSN:   getEnv("SENTRY_DSN", ""),
		},

		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "user"),
			Password:        getEnv("DB_PASSWORD", "password"),
			Name:            getEnv("DB_NAME", "genshin_quiz"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", "5m"),
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

	logger, err := app.initializeLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	app.Logger = logger

	app.Logger.Info("Current App Config", zap.Any("config", app.Config))

	err = app.initializeSentry()
	if err != nil {
		app.Logger.Error("Failed to initialize Sentry", zap.Error(err))
		// 不要因为 Sentry 初始化失败而崩溃应用
	}

	db, err := app.initializeDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	app.DB = db

	return app
}
