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

func GetVote(
	ctx context.Context,
	app *config.App,
	req oapi.GetVoteRequestObject,
) (*oapi.GetVote200JSONResponse, error) {
	// 调用仓库层获取投票详情（使用默认语言）
	res, err := vote_repo.GetVoteByUUID(ctx, app.DB, req.Id, nil)
	if err != nil {
		return nil, err
	}
	voteDBId := res.Vote.ID

	// 获取选项
	options, err := vote_repo.GetVoteOptions(ctx, app.DB, voteDBId)
	if err != nil {
		return nil, err
	}

	// 获取选项ID列表
	ids := make([]int64, 0, len(*options))
	for _, opt := range *options {
		ids = append(ids, opt.ID)
	}

	// 获取选项翻译
	optionTranslations, err := vote_repo.GetVoteOptionTranslations(ctx, app.DB, ids, nil)
	if err != nil {
		return nil, err
	}

	// 构建 DetailedVote
	detailedVote := dao.DetailedVote{
		Vote:               res.Vote,
		User:               res.User,
		Translation:        res.Translation,
		Options:            *options,
		OptionTranslations: *optionTranslations,
	}

	// 检查用户是否已登录，如果已登录则获取用户的投票记录和点赞状态
	votedOptions := make(map[string]int)
	likeStatus := int16(0)
	userClaims, ok := middleware.GetUserFromContextOnly(ctx)
	if ok {
		// 获取用户已投票的选项
		userVotes, err := vote_repo.GetUserVotedOptions(ctx, app.DB, userClaims.UserID, voteDBId)
		if err != nil {
			return nil, err
		}

		// 构建选项UUID到投票数的映射
		optionIDToUUID := make(map[int64]string)
		for _, opt := range *options {
			optionIDToUUID[opt.ID] = opt.OptionUUID.String()
		}

		for _, uv := range *userVotes {
			if optionUUID, exists := optionIDToUUID[uv.OptionID]; exists {
				votedOptions[optionUUID] = int(uv.VoteCount)
			}
		}

		// 获取用户的点赞状态
		userLikeStatus, err := vote_repo.GetVoteLikeStatus(
			ctx,
			app.DB,
			userClaims.UserID,
			voteDBId,
		)
		if err != nil {
			return nil, err
		}
		// 如果有点赞记录，使用实际值；否则保持默认值0
		if userLikeStatus != nil {
			likeStatus = *userLikeStatus
		}
	}

	// 转换为 DTO
	dto := transformer.ConvertDetailedVoteToDTO(detailedVote, votedOptions, likeStatus)

	response := oapi.GetVote200JSONResponse(dto)
	return &response, nil
}
