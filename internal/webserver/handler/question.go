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
	return *res, nil
}

func (h *Handler) GetQuestion(
	ctx context.Context,
	req oapi.GetQuestionRequestObject,
) (oapi.GetQuestionResponseObject, error) {
	res, err := services.GetQuestion(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return (oapi.GetQuestion200JSONResponse)(*res), nil
}

func (h *Handler) PostCreateQuestion(
	ctx context.Context,
	req oapi.PostCreateQuestionRequestObject,
) (oapi.PostCreateQuestionResponseObject, error) {
	res, err := services.PostCreateQuestion(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (h *Handler) GetQuestionMySubmissions(
	ctx context.Context,
	req oapi.GetQuestionMySubmissionsRequestObject,
) (oapi.GetQuestionMySubmissionsResponseObject, error) {
	res, err := services.GetQuestionMySubmissions(ctx, h.app, req)
	if err != nil {
		return nil, err
	}

	return (oapi.GetQuestionMySubmissions200JSONResponse)(*res), nil
}

func (h *Handler) GetQuestionRecentSubmissions(
	ctx context.Context,
	req oapi.GetQuestionRecentSubmissionsRequestObject,
) (oapi.GetQuestionRecentSubmissionsResponseObject, error) {
	res, err := services.GetQuestionRecentSubmissions(ctx, h.app, req)
	if err != nil {
		return nil, err
	}

	return (oapi.GetQuestionRecentSubmissions200JSONResponse)(*res), nil
}

func (h *Handler) PostSubmitAnswer(
	ctx context.Context,
	req oapi.PostSubmitAnswerRequestObject,
) (oapi.PostSubmitAnswerResponseObject, error) {
	res, err := services.PostSubmitAnswer(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (h *Handler) PostLikeQuestion(
	ctx context.Context,
	req oapi.PostLikeQuestionRequestObject,
) (oapi.PostLikeQuestionResponseObject, error) {
	err := services.PostLikeQuestion(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return oapi.PostLikeQuestion201Response{}, nil
}
