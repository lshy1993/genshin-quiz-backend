package handler

import (
	"context"
	"genshin-quiz/generated/oapi"
	poll_services "genshin-quiz/internal/services/poll"
	question_services "genshin-quiz/internal/services/question"
	services "genshin-quiz/internal/services/user"
)

func (h *Handler) GetUsers(
	ctx context.Context,
	req oapi.GetUsersRequestObject,
) (oapi.GetUsersResponseObject, error) {
	res, err := services.GetUsers(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return *res, nil
}

func (h *Handler) GetUser(
	ctx context.Context,
	req oapi.GetUserRequestObject,
) (oapi.GetUserResponseObject, error) {
	res, err := services.GetUser(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return (oapi.GetUser200JSONResponse)(*res), nil
}

func (h *Handler) GetCurrentUser(
	ctx context.Context,
	req oapi.GetCurrentUserRequestObject,
) (oapi.GetCurrentUserResponseObject, error) {
	res, err := services.GetMe(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return (oapi.GetCurrentUser200JSONResponse)(*res), nil
}

func (h *Handler) UpdateUser(
	ctx context.Context,
	req oapi.UpdateUserRequestObject,
) (oapi.UpdateUserResponseObject, error) {
	res, err := services.UpdateUser(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return (oapi.UpdateUser200JSONResponse)(*res), nil
}

func (h *Handler) GetUserPolls(
	ctx context.Context,
	req oapi.GetUserPollsRequestObject,
) (oapi.GetUserPollsResponseObject, error) {
	res, err := poll_services.GetPollsByUser(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return (oapi.GetUserPolls200JSONResponse)(*res), nil
}

func (h *Handler) GetUserQuestions(
	ctx context.Context,
	req oapi.GetUserQuestionsRequestObject,
) (oapi.GetUserQuestionsResponseObject, error) {
	res, err := question_services.GetQuestionsByUser(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return (oapi.GetUserQuestions200JSONResponse)(*res), nil
}
