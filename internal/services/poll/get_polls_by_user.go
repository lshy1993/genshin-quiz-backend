package services

import (
	"context"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/common"
	"genshin-quiz/internal/dao"
	poll_repo "genshin-quiz/internal/repository/poll"
	"genshin-quiz/internal/webserver/middleware"
)

func GetPollsByUser(
	ctx context.Context,
	app *config.App,
	req oapi.GetUserPollsRequestObject,
) (*oapi.GetUserPolls200JSONResponse, error) {
	// 从 context 中获取登录用户信息(允许为空)
	userClaims, ok := middleware.GetUserFromContextOnly(ctx)
	if !ok {
		return nil, common.ErrUserNotInContext
	}

	page := 1
	if req.Params.Page != nil {
		page = *req.Params.Page
	}

	limit := 25
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}

	// 调用 repository 层获取数据
	param := dao.PollListParams{
		Page:       page,
		NumPerPage: limit,
		Author:     &userClaims.UserID,
		Type:       "all",
		SortBy:     "",
		SortDesc:   false,
	}
	result, err := poll_repo.GetPolls(ctx, app.DB, param)
	if err != nil {
		return nil, err
	}

	dtos, err := buildPollsWithLike(ctx, app.DB, result)

	return &oapi.GetUserPolls200JSONResponse{
		Polls: dtos,
		Total: result.Total,
	}, nil
}
