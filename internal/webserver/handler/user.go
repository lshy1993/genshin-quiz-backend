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
	return (oapi.GetUsers200JSONResponse{
		Limit:  ptr(0),
		Offset: ptr(0),
		Total:  ptr(0),
		Users:  &[]oapi.User{},
	}), nil
}
