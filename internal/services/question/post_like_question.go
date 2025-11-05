package services

import (
	"context"
	"fmt"

	"genshin-quiz/config"
	question_repo "genshin-quiz/internal/repository/question"
	"genshin-quiz/internal/webserver/middleware"

	"github.com/google/uuid"
)

func PostLikeQuestion(
	ctx context.Context,
	app *config.App,
	questionUUID uuid.UUID,
	value int16, // 1=点赞, -1=点踩, 0=取消
) error {
	userClaims, ok := middleware.GetUserFromContextOnly(ctx)
	if !ok {
		return fmt.Errorf("user not found in context")
	}

	// 验证 value 参数
	if value != -1 && value != 0 && value != 1 {
		return fmt.Errorf("invalid like value: %d, must be -1, 0, or 1", value)
	}

	// 更新点赞状态
	return question_repo.UpsertQuestionLike(ctx, app.DB, userClaims.UserID, questionUUID, value)
}
