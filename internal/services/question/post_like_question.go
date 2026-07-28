package services

import (
	"context"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/common"
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
		return common.ErrUserNotInContext
	}

	questionUUID := req.Id
	value := int16(req.Body.Like)

	questionInfo, err := question_repo.GetQuestionByUUID(ctx, app.DB, questionUUID)
	if err != nil {
		return err
	}

	qDbID := questionInfo.Question.ID

	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 更新点赞状态
	err = question_repo.UpsertQuestionLike(ctx, tx, userClaims.UserID, qDbID, value)
	if err != nil {
		return err
	}
	// 更新问题的点赞数
	err = question_repo.UpdateQuestionLikeCount(ctx, tx, qDbID, value)
	if err != nil {
		return err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
