package handler

import (
	"context"
	"genshin-quiz/generated/oapi"
	services "genshin-quiz/internal/services/user"
)

func (h *Handler) PostRegisterUser(
	ctx context.Context,
	req oapi.PostRegisterUserRequestObject,
) (oapi.PostRegisterUserResponseObject, error) {
	res, err := services.RegisterUser(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return (oapi.PostRegisterUser201JSONResponse)(*res), nil
}

func (h *Handler) PostLoginUser(
	ctx context.Context,
	req oapi.PostLoginUserRequestObject,
) (oapi.PostLoginUserResponseObject, error) {
	res, err := services.LoginUser(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return (oapi.PostLoginUser200JSONResponse)(*res), nil
}

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
