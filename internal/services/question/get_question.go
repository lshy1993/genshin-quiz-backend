package services

import (
	"context"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	dao "genshin-quiz/internal/dao"
	"genshin-quiz/internal/dao/transformer"
	question_repo "genshin-quiz/internal/repository/question"
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
	// 获取选项
	options, err := question_repo.GetQuestionOptions(ctx, app.DB, res.Question.ID)
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
	count, err := question_repo.GetQuestionSubmissionCount(ctx, app.DB, res.Question.ID)
	if err != nil {
		return nil, err
	}

	dto := transformer.ConvertDetailToQuestion(dao.DetailedQuestion{
		Question:           res.Question,
		User:               res.User,
		Translation:        res.Translation,
		Options:            *options,
		OptionTranslations: *optionTranslations,
		SubmissionCount:    *count,
	})

	return &dto, nil
}
