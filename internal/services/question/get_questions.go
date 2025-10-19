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
	var page int
	if req.Params.Page != nil {
		page = *req.Params.Page
	} else {
		page = 0
	}
	var limit int
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	} else {
		limit = 25
	}

	sortDesc := false
	if req.Params.SortDesc != nil {
		sortDesc = *req.Params.SortDesc
	}

	param := dao.QuestionListParams{
		Page:       page,
		NumPerPage: limit,
		Category:   req.Params.Category,
		Difficulty: req.Params.Difficulty,
		Query:      req.Params.Query,
		SortBy:     req.Params.SortBy,
		SortDesc:   sortDesc,
		Language:   req.Params.Language,
	}

	dao, err := question_repo.GetQuestions(ctx, app.DB, param)
	if err != nil {
		return nil, err
	}

	return &oapi.GetQuestions200JSONResponse{
		Questions: dao.Questions,
		Total:     dao.Total,
	}, nil
}
