package services

import (
	"context"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/dao"
	poll_repo "genshin-quiz/internal/repository/poll"
	user_repo "genshin-quiz/internal/repository/user"
)

func GetPollsByUser(
	ctx context.Context,
	app *config.App,
	req oapi.GetUserPollsRequestObject,
) (*oapi.GetUserPolls200JSONResponse, error) {
	userInfo, err := user_repo.GetUserInfoByUUID(ctx, app.DB, req.Id)
	if err != nil {
		return nil, err
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
		Author:     &userInfo.ID,
		Type:       "all",
		SortBy:     "",
		SortDesc:   false,
	}
	result, err := poll_repo.GetPolls(ctx, app.DB, param)
	if err != nil {
		return nil, err
	}

	dtos, err := poll_repo.BuildPollsWithLike(ctx, app.DB, result)
	if err != nil {
		return nil, err
	}

	return &oapi.GetUserPolls200JSONResponse{
		Polls: dtos,
		Total: result.Total,
	}, nil
}
