package handler

import (
	"context"
	"genshin-quiz/generated/oapi"
)

func (h *Handler) GetUsers(ctx context.Context, req oapi.GetUsersRequestObject) (oapi.GetUsersResponseObject, error) {
	return (oapi.GetUsers200JSONResponse{
		Limit:  ptr(0),
		Offset: ptr(0),
		Total:  ptr(0),
		Users:  &[]oapi.User{},
	}), nil
}
