package cronjob

import (
	"context"
	"time"

	"genshin-quiz/config"
	question_repo "genshin-quiz/internal/repository/question"
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

// RecalibrateQuestionStats 定期校准每个问题的统计数据（包括点赞和提交统计）
func (c *Cronjob) RecalibrateQuestionStats() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	c.app.Logger.Info("Starting question statistics recalibration...")

	// 更新点赞统计
	err := question_repo.RecalculateAllQuestionLikeStats(ctx, c.app.DB)
	if err != nil {
		c.app.Logger.Error("Failed to recalibrate question like stats: " + err.Error())
		return err
	}

	// 更新提交统计（正确率等）
	err = question_repo.RecalculateAllQuestionSubmissionStats(ctx, c.app.DB)
	if err != nil {
		c.app.Logger.Error("Failed to recalibrate question submission stats: " + err.Error())
		return err
	}

	c.app.Logger.Info("Question statistics recalibration completed successfully")
	return nil
}
