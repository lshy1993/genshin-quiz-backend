package services

import (
	"context"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	dao "genshin-quiz/internal/dao"
	"genshin-quiz/internal/dao/transformer"
	question_repo "genshin-quiz/internal/repository/question"
	"genshin-quiz/internal/webserver/middleware"
)

func GetQuestion(
	ctx context.Context,
	app *config.App,
	req oapi.GetQuestionRequestObject,
) (*oapi.Question, error) {
	// 调用仓库层获取问题详情
	res, err := question_repo.GetQuestionByUUID(ctx, app.DB, req.Id, nil)
	if err != nil {
		return nil, err
	}
	questionDBId := res.Question.ID
	// 获取选项
	options, err := question_repo.GetQuestionOptions(ctx, app.DB, questionDBId)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(*options))
	for _, opt := range *options {
		ids = append(ids, opt.ID)
	}
	// 获取选项翻译
	optionTranslations, err := question_repo.GetQuestionOptionTranslations(ctx, app.DB, ids, nil)
	if err != nil {
		return nil, err
	}
	// 获取统计信息
	count, err := question_repo.GetQuestionSubmissionCount(ctx, app.DB, questionDBId)
	if err != nil {
		return nil, err
	}

	// 构建 DetailedQuestion
	detailedQuestion := dao.DetailedQuestion{
		Question:           res.Question,
		User:               res.User,
		Translation:        res.Translation,
		Options:            *options,
		OptionTranslations: *optionTranslations,
		SubmissionCount:    *count,
	}

	// 检查用户是否已解答此题（如果用户已登录）
	solved := false
	likeStatus := int16(0)
	userClaims, ok := middleware.GetUserFromContextOnly(ctx)
	if ok {
		// 检查是否已解答
		solved, err = question_repo.CheckQuestionSolved(
			ctx,
			app.DB,
			userClaims.UserID,
			questionDBId,
		)
		if err != nil {
			return nil, err
		}

		// 检查用户的点赞状态
		userLikeStatus, err := question_repo.GetQuestionLikeStatus(
			ctx,
			app.DB,
			userClaims.UserID,
			questionDBId,
		)
		if err != nil {
			return nil, err
		}
		// 如果有点赞记录，使用实际值；否则保持默认值0
		if userLikeStatus != nil {
			likeStatus = *userLikeStatus
		}
	}

	dto := transformer.ConvertDetailToQuestion(detailedQuestion, solved, likeStatus)

	return &dto, nil
}
