package handler

import (
	"context"
	"genshin-quiz/generated/oapi"
)

func GetQuizzes(ctx context.Context, req oapi.GetQuizzesRequestObject) (oapi.GetQuizzesResponseObject, error) {
	return (oapi.GetQuizzes200JSONResponse{}), nil
}
