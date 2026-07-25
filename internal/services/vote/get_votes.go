package services

import (
	"context"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/dao"
	"genshin-quiz/internal/dao/transformer"
	vote_repo "genshin-quiz/internal/repository/vote"
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

	voteType := "all"
	if req.Params.Type != nil {
		voteType = string(*req.Params.Type)
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
	result, err := vote_repo.GetVotes(
		ctx,
		app.DB,
		dao.VoteListParams{
			Page:     page,
			Limit:    limit,
			Type:     voteType,
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
	voteIDs := make([]int64, 0, len(result.Votes))
	for _, vote := range result.Votes {
		voteIDs = append(voteIDs, vote.Vote.ID)
	}

	// 批量获取所有投票的点赞数
	likesCountMap, err := vote_repo.GetMultipleVotesLikesCount(ctx, app.DB, voteIDs)
	if err != nil {
		return nil, err
	}

	// 转换为 DTO
	dtos := make([]oapi.Poll, 0, len(result.Votes))
	for _, vote := range result.Votes {
		// 覆盖投票的点赞数为实时计算的值
		vote.Vote.LikesCount = likesCountMap[vote.Vote.ID]

		voted := false
		likeStatus := int16(0)

		// 如果用户已登录，检查投票状态和点赞状态
		if userClaims != nil {
			// 检查用户是否已投票
			userVotes, err := vote_repo.GetUserVotedOptions(
				ctx,
				app.DB,
				userClaims.UserID,
				vote.Vote.ID,
			)
			if err == nil && userVotes != nil && len(*userVotes) > 0 {
				voted = true
			}

			// 获取点赞状态
			userLikeStatus, err := vote_repo.GetVoteLikeStatus(
				ctx,
				app.DB,
				userClaims.UserID,
				vote.Vote.ID,
			)
			if err == nil && userLikeStatus != nil {
				likeStatus = *userLikeStatus
			}
		}

		dto := transformer.ConvertSimpleVoteToDTO(vote, voted, likeStatus)
		dtos = append(dtos, dto)
	}

	return &oapi.GetPolls200JSONResponse{
		Total: result.Total,
		Polls: dtos,
	}, nil
}
