package handler

import (
	"context"

	"genshin-quiz/generated/oapi"
	home_services "genshin-quiz/internal/services/home"
)

func (h *Handler) GetHome(
	ctx context.Context,
	req oapi.GetHomeRequestObject,
) (oapi.GetHomeResponseObject, error) {
	res, err := home_services.GetHome(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return *res, nil
}
