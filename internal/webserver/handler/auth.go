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

func (h *Handler) PostForgotPassword(
	ctx context.Context,
	req oapi.PostForgotPasswordRequestObject,
) (oapi.PostForgotPasswordResponseObject, error) {
	res, err := services.ForgotPassword(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return *res, nil
}

func (h *Handler) PostChangePassword(
	ctx context.Context,
	req oapi.PostChangePasswordRequestObject,
) (oapi.PostChangePasswordResponseObject, error) {
	res, err := services.ChangePassword(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return *res, nil
}

func (h *Handler) PostResetPassword(
	ctx context.Context,
	req oapi.PostResetPasswordRequestObject,
) (oapi.PostResetPasswordResponseObject, error) {
	res, err := services.ResetPassword(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return *res, nil
}

func (h *Handler) PostSendVerificationEmail(
	ctx context.Context,
	req oapi.PostSendVerificationEmailRequestObject,
) (oapi.PostSendVerificationEmailResponseObject, error) {
	res, err := services.SendVerificationEmail(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return *res, nil
}

func (h *Handler) PostVerifyEmail(
	ctx context.Context,
	req oapi.PostVerifyEmailRequestObject,
) (oapi.PostVerifyEmailResponseObject, error) {
	res, err := services.VerifyEmail(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return *res, nil
}
