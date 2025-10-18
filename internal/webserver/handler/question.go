package handler

import (
	"context"
	"genshin-quiz/generated/oapi"
	services "genshin-quiz/internal/services/question"
)

func (h *Handler) GetQuestions(
	ctx context.Context,
	req oapi.GetQuestionsRequestObject,
) (oapi.GetQuestionsResponseObject, error) {
	res, err := services.GetQuestions(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return oapi.GetQuestions200JSONResponse(*res), nil
}
