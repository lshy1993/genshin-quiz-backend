package handler

import (
	"context"
	"genshin-quiz/generated/oapi"
)

func GetQuestions(ctx context.Context, req oapi.GetQuestionsRequestObject) (oapi.GetQuestionsResponseObject, error) {
	return (oapi.GetQuestions200JSONResponse{}), nil
}
