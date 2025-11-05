package cronjob

import (
	"context"
	"time"

	"genshin-quiz/config"
	user_repo "genshin-quiz/internal/repository/user"
)

type Cronjob struct {
	app *config.App
}

func NewCronjob(app *config.App) *Cronjob {
	return &Cronjob{
		app: app,
	}
}

// RecalibrateUserStats 定期校准用户统计数据
// 这个任务用于修正因为并发、异常等原因导致的统计数据不一致.
func (c *Cronjob) RecalibrateUserStats() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	c.app.Logger.Info("Starting user statistics recalibration...")

	err := user_repo.RecalculateAllUserStats(ctx, c.app.DB)
	if err != nil {
		c.app.Logger.Error("Failed to recalibrate user stats: " + err.Error())
		return err
	}

	c.app.Logger.Info("User statistics recalibration completed successfully")
	return nil
}
