package handler

import (
	"context"
	"genshin-quiz/generated/oapi"
)

func (h *Handler) GetQuizzes(ctx context.Context, req oapi.GetQuizzesRequestObject) (oapi.GetQuizzesResponseObject, error) {
	return (oapi.GetQuizzes200JSONResponse{
		Limit:   ptr(0),
		Offset:  ptr(0),
		Quizzes: &[]oapi.Quiz{},
		Total:   ptr(0),
	}), nil
}
