package handler

import (
	"context"
	"genshin-quiz/generated/oapi"
)

func (*Handler) GetQuestions(
	ctx context.Context,
	req oapi.GetQuestionsRequestObject,
) (oapi.GetQuestionsResponseObject, error) {
	return (oapi.GetQuestions200JSONResponse{}), nil
}
