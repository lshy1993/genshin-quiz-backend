package services

import (
	"context"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	question_repo "genshin-quiz/internal/repository/question"
)

func GetQuestion(
	ctx context.Context,
	app *config.App,
	req oapi.GetQuestionRequestObject,
) (*oapi.GetQuestion200JSONResponse, error) {
	res, err := question_repo.GetQuestionByUUID(ctx, app.DB, req.Id)
	if err != nil {
		return nil, err
	}

	resp := oapi.GetQuestion200JSONResponse(*res)
	return &resp, nil
}
