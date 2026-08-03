package services

import (
	"context"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/dao"
	poll_repo "genshin-quiz/internal/repository/poll"
)

func GetPolls(
	ctx context.Context,
	app *config.App,
	req oapi.GetPollsRequestObject,
) (*oapi.GetPolls200JSONResponse, error) {
	// 设置默认值
	page := 1
	if req.Params.Page != nil {
		page = *req.Params.Page
	}

	limit := 25
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}

	pollType := "all"
	if req.Params.Type != nil {
		pollType = string(*req.Params.Type)
	}

	sortBy := "created_at"
	if req.Params.SortBy != nil {
		sortBy = *req.Params.SortBy
	}

	sortDesc := false
	if req.Params.SortDesc != nil {
		sortDesc = *req.Params.SortDesc
	}

	// 调用 repository 层获取数据
	param := dao.PollListParams{
		Page:       page,
		NumPerPage: limit,
		Type:       pollType,
		Query:      req.Params.Query,
		Language:   req.Params.Language,
		SortBy:     sortBy,
		SortDesc:   sortDesc,
	}
	result, err := poll_repo.GetPolls(ctx, app.DB, param)
	if err != nil {
		return nil, err
	}

	dtos, err := poll_repo.BuildPollsWithLike(ctx, app.DB, result)
	if err != nil {
		return nil, err
	}

	return &oapi.GetPolls200JSONResponse{
		Total: result.Total,
		Polls: dtos,
	}, nil
}
