package services

import (
	"context"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/dao"
	"genshin-quiz/internal/dao/transformer"
	poll_repo "genshin-quiz/internal/repository/poll"
	"genshin-quiz/internal/webserver/middleware"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

func GetPoll(
	ctx context.Context,
	app *config.App,
	req oapi.GetPollRequestObject,
) (*oapi.GetPoll200JSONResponse, error) {
	// 调用仓库层获取投票详情（使用默认语言）
	res, err := poll_repo.GetPollByUUID(ctx, app.DB, req.Id, nil)
	if err != nil {
		return nil, err
	}
	pollDBId := res.Poll.ID

	// 获取选项
	options, err := poll_repo.GetPollOptions(ctx, app.DB, pollDBId)
	if err != nil {
		return nil, err
	}

	// 获取选项ID列表
	ids := make([]int64, 0, len(*options))
	for _, opt := range *options {
		ids = append(ids, opt.ID)
	}

	// 获取选项翻译
	optionTranslations, err := poll_repo.GetPollOptionTranslations(ctx, app.DB, ids, nil)
	if err != nil {
		return nil, err
	}

	// 构建 DetailedPoll
	detailedPoll := dao.DetailedPoll{
		Poll:               res.Poll,
		User:               res.User,
		Translation:        res.Translation,
		Options:            *options,
		OptionTranslations: *optionTranslations,
	}

	// 获取投票的实时点赞数
	likesCount, err := poll_repo.GetPollLikesCount(ctx, app.DB, pollDBId)
	if err != nil {
		return nil, err
	}
	detailedPoll.Poll.LikesCount = likesCount

	// 检查用户是否已登录，如果已登录则获取用户的投票记录和点赞状态
	polldOptions := []oapi.PollVote{}
	likeStatus := int16(0)
	userClaims, ok := middleware.GetUserFromContextOnly(ctx)
	if ok {
		// 获取用户已投票的选项
		userVotes, err := poll_repo.GetUserVotedOptions(ctx, app.DB, userClaims.UserID, pollDBId)
		if err != nil {
			return nil, err
		}

		// 构建选项ID到选项UUID的映射
		optionIDToUUID := make(map[int64]openapi_types.UUID)
		for _, opt := range *options {
			optionIDToUUID[opt.ID] = opt.OptionUUID
		}

		// 构建VoteSubmissionOption数组
		for _, uv := range *userVotes {
			if optionUUID, exists := optionIDToUUID[uv.OptionID]; exists {
				polldOptions = append(polldOptions, oapi.PollVote{
					OptionId: optionUUID,
					Votes:    int(uv.VoteCount),
				})
			}
		}

		// 获取用户的点赞状态
		userLikeStatus, err := poll_repo.GetPollLikeStatus(
			ctx,
			app.DB,
			userClaims.UserID,
			pollDBId,
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
	dto := transformer.ConvertDetailedVoteToDTO(detailedPoll, polldOptions, likeStatus)

	response := oapi.GetPoll200JSONResponse(dto)
	return &response, nil
}
