package console

import (
	"fmt"

	"genshin-quiz/config"
)

type Console struct {
	app *config.App
}

func NewConsole(app *config.App) *Console {
	return &Console{
		app: app,
	}
}

func (c *Console) Execute(command string, args []string) error {
	switch command {
	case "seed":
		return c.SeedDatabase()
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func (c *Console) SeedDatabase() error {
	c.app.Logger.Info("Seeding database with sample data...")

	// Create sample users
	// sampleUsers := []model.Users{
	// 	{
	// 		Username: "admin",
	// 		Email:    "admin@genshinquiz.com",
	// 		Password: "admin123",
	// 	},
	// 	{
	// 		Username: "testuser1",
	// 		Email:    "user1@example.com",
	// 		Password: "password123",
	// 	},
	// 	{
	// 		Username: "testuser2",
	// 		Email:    "user2@example.com",
	// 		Password: "password123",
	// 	},
	// }
	// for _, userReq := range sampleUsers {
	// }
	return nil
}
