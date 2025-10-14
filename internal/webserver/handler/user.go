package handler

import (
	"context"
	"genshin-quiz/generated/oapi"
)

func (*Handler) GetUsers(ctx context.Context, req oapi.GetUsersRequestObject) (oapi.GetUsersResponseObject, error) {
	return (oapi.GetUsers200JSONResponse{}), nil
}
