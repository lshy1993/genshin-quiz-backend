package services

import (
	"context"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/common"
	question_repo "genshin-quiz/internal/repository/question"
	"genshin-quiz/internal/webserver/middleware"
)

func PostLikeVote(
	ctx context.Context,
	app *config.App,
	req oapi.PostLikeVoteRequestObject,
) error {
	userClaims, ok := middleware.GetUserFromContextOnly(ctx)
	if !ok {
		return common.ErrUserNotInContext
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
