package services

import (
	"context"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	dao "genshin-quiz/internal/dao"
	question_repo "genshin-quiz/internal/repository/question"
)

func GetQuestions(
	ctx context.Context,
	app *config.App,
	req oapi.GetQuestionsRequestObject,
) (*oapi.GetQuestions200JSONResponse, error) {
	page := 1 // 修正：页码从1开始，不是0
	if req.Params.Page != nil {
		page = *req.Params.Page
	}

	limit := 25
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}
	sortBy := "PublishDate"
	if req.Params.SortBy != nil {
		sortBy = *req.Params.SortBy
	}
	sortDesc := false
	if req.Params.SortDesc != nil {
		sortDesc = *req.Params.SortDesc
	}

	// 调用 repository 层获取数据
	param := dao.QuestionListParams{
		Page:       page,
		NumPerPage: limit,
		Category:   req.Params.Category,
		Difficulty: req.Params.Difficulty,
		Query:      req.Params.Query,
		Language:   req.Params.Language,
		SortBy:     sortBy,
		SortDesc:   sortDesc,
	}
	result, err := question_repo.GetQuestions(ctx, app.DB, param)
	if err != nil {
		return nil, err
	}

	dtos, err := question_repo.BuildQuestionsWithTransaction(ctx, app.DB, result)
	if err != nil {
		return nil, err
	}

	return &oapi.GetQuestions200JSONResponse{
		Questions: dtos,
		Total:     result.Total,
	}, nil
}
