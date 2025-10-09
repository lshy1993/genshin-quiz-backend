package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"genshin-quiz/config"
	"genshin-quiz/internal/webserver"
)

func main() {
	// Load environment variables based on environment
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development" // 默认为开发环境
	}

	var envFile string
	switch env {
	case "development":
		envFile = ".env.dev"
	case "testing":
		envFile = ".env.test"
	case "production":
		envFile = ".env.prod"
	default:
		envFile = ".env.dev"
	}
	// Load variables from the specified file
	if err := godotenv.Load(envFile); err != nil {
		log.Printf("Warning: Error loading %s file: %v", envFile, err)
	}

	// Initialize configuration
	app := config.NewApp()

	// Initialize server
	server := webserver.NewServer(app)
	// Start server in a goroutine
	server.Start()
}
