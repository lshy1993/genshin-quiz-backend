package services

import (
	"context"
	"fmt"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	question_repo "genshin-quiz/internal/repository/question"
	"genshin-quiz/internal/webserver/middleware"
)

func PostLikeQuestion(
	ctx context.Context,
	app *config.App,
	req oapi.PostLikeQuestionRequestObject,
) error {
	userClaims, ok := middleware.GetUserFromContextOnly(ctx)
	if !ok {
		return fmt.Errorf("user not found in context")
	}

	questionUUID := req.Id
	value := int16(req.Body.Like)

	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// 更新点赞状态
	err = question_repo.UpsertQuestionLike(ctx, tx, userClaims.UserID, questionUUID, value)
	if err != nil {
		tx.Rollback()
		return err
	}
	// 更新问题的点赞数
	err = question_repo.UpdateQuestionLikeCount(ctx, tx, questionUUID, value)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
