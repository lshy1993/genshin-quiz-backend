package services

import (
	"context"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/dao"
	"genshin-quiz/internal/dao/transformer"
	poll_repo "genshin-quiz/internal/repository/poll"
	"genshin-quiz/internal/webserver/middleware"
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
	result, err := poll_repo.GetPolls(
		ctx,
		app.DB,
		dao.PollListParams{
			Page:     page,
			Limit:    limit,
			Type:     pollType,
			Query:    req.Params.Query,
			Language: req.Params.Language,
			SortBy:   sortBy,
			SortDesc: sortDesc,
		},
	)
	if err != nil {
		return nil, err
	}

	// 获取当前用户信息（可选）
	var userClaims *middleware.UserClaims
	if claims, ok := middleware.GetUserFromContextOnly(ctx); ok {
		userClaims = claims
	}

	// 提取所有投票的ID用于批量查询
	pollIDs := make([]int64, 0, len(result.Polls))
	for _, poll := range result.Polls {
		pollIDs = append(pollIDs, poll.Poll.ID)
	}

	// 批量获取所有投票的点赞数
	likesCountMap, err := poll_repo.GetMultiplePollsLikesCount(ctx, app.DB, pollIDs)
	if err != nil {
		return nil, err
	}

	// 转换为 DTO
	dtos := make([]oapi.Poll, 0, len(result.Polls))
	for _, poll := range result.Polls {
		// 覆盖投票的点赞数为实时计算的值
		poll.Poll.LikesCount = likesCountMap[poll.Poll.ID]

		voted := false
		likeStatus := int16(0)

		// 如果用户已登录，检查投票状态和点赞状态
		if userClaims != nil {
			// 检查用户是否已投票
			userVotes, err := poll_repo.GetUserVotedOptions(
				ctx,
				app.DB,
				userClaims.UserID,
				poll.Poll.ID,
			)
			if err == nil && userVotes != nil && len(*userVotes) > 0 {
				voted = true
			}

			// 获取点赞状态
			userLikeStatus, err := poll_repo.GetPollLikeStatus(
				ctx,
				app.DB,
				userClaims.UserID,
				poll.Poll.ID,
			)
			if err == nil && userLikeStatus != nil {
				likeStatus = *userLikeStatus
			}
		}

		dto := transformer.ConvertSimplePollToDTO(poll, voted, likeStatus)
		dtos = append(dtos, dto)
	}

	return &oapi.GetPolls200JSONResponse{
		Total: result.Total,
		Polls: dtos,
	}, nil
}
