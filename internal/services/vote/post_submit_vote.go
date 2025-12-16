package services

import (
	"context"
	"fmt"
	"time"

	"genshin-quiz/config"
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/oapi"
	vote_repo "genshin-quiz/internal/repository/vote"
	"genshin-quiz/internal/webserver/middleware"

	"github.com/google/uuid"
)

func PostVote(
	ctx context.Context,
	app *config.App,
	req oapi.PostVoteRequestObject,
) (oapi.PostVoteResponseObject, error) {
	// 从 context 中获取用户信息
	userClaims, ok := middleware.GetUserFromContextOnly(ctx)
	if !ok {
		return nil, fmt.Errorf("user not found in context")
	}

	// 获取投票信息
	voteInfo, err := vote_repo.GetVoteByUUID(ctx, app.DB, req.Id, nil)
	if err != nil {
		return nil, fmt.Errorf("vote not found: %w", err)
	}

	// 检查投票是否已过期
	if voteInfo.Vote.ExpiresAt != nil && voteInfo.Vote.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("vote has expired")
	}

	// 检查投票是否还未开始
	if voteInfo.Vote.StartAt.After(time.Now()) {
		return nil, fmt.Errorf("vote has not started yet")
	}

	// 获取所有选项
	options, err := vote_repo.GetVoteOptions(ctx, app.DB, voteInfo.Vote.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get vote options: %w", err)
	}

	// 构建选项UUID到ID的映射
	optionUUIDToID := make(map[uuid.UUID]int64)
	for _, opt := range *options {
		optionUUIDToID[opt.OptionUUID] = opt.ID
	}

	// 计算总投票数并验证
	totalVotes := 0
	optionVotes := make(map[int64]int32) // optionID -> votes

	for _, optionVote := range req.Body.Options {
		optionID, exists := optionUUIDToID[optionVote.OptionId]
		if !exists {
			return nil, fmt.Errorf("invalid option id: %s", optionVote.OptionId)
		}

		totalVotes += optionVote.Votes
		optionVotes[optionID] = int32(optionVote.Votes)

		// 检查每个选项的投票数是否超过限制
		if voteInfo.Vote.VotesPerOption != nil && *voteInfo.Vote.VotesPerOption > 0 {
			if optionVote.Votes > int(*voteInfo.Vote.VotesPerOption) {
				return nil, fmt.Errorf("votes per option exceeded limit")
			}
		}
	}

	// 检查总投票数是否超过用户限制
	if totalVotes > int(voteInfo.Vote.VotesPerUser) {
		return nil, fmt.Errorf("total votes exceeded limit: %d > %d", totalVotes, voteInfo.Vote.VotesPerUser)
	}

	// 检查用户是否已经投过票
	existingVotes, err := vote_repo.GetUserVotedOptions(ctx, app.DB, userClaims.UserID, voteInfo.Vote.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user votes: %w", err)
	}

	// 如果用户已经投过票，返回错误（或者可以选择更新投票）
	if len(*existingVotes) > 0 {
		return nil, fmt.Errorf("user has already voted")
	}

	// 开始事务
	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()

	// 插入用户投票记录
	for optionID, voteCount := range optionVotes {
		userVote := model.UserVotes{
			VoteID:    voteInfo.Vote.ID,
			UserID:    userClaims.UserID,
			OptionID:  optionID,
			VoteCount: voteCount,
			CreatedAt: now,
			UpdatedAt: now,
		}

		err = vote_repo.UpsertUserVote(ctx, tx, userVote)
		if err != nil {
			return nil, fmt.Errorf("failed to insert user vote: %w", err)
		}
	}

	// 更新投票统计（参与人数和总票数）
	err = vote_repo.RecalculateVoteStats(ctx, tx, voteInfo.Vote.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update vote stats: %w", err)
	}

	// 重新计算选项统计
	err = vote_repo.RecalculateVoteOptionStats(ctx, tx, voteInfo.Vote.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update option stats: %w", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return oapi.PostVote200Response{}, nil
}
