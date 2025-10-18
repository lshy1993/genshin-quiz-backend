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

	param := dao.QuestionListParams{
		Page:       *req.Params.Page,
		NumPerPage: *req.Params.Limit,
		Category:   string(*req.Params.Category),
		Difficulty: string(*req.Params.Difficulty),
		Query:      *req.Params.Query,
		SortBy:     *req.Params.SortBy,
		SortDesc:   *req.Params.SortDesc,
		Language:   *req.Params.Language,
	}

	dao, err := question_repo.GetQuestions(ctx, app.DB, param)
	if err != nil {
		return nil, err
	}

	return &oapi.GetQuestions200JSONResponse{
		Questions: dao.Questions,
		Total:     int(dao.Total),
	}, nil
}
