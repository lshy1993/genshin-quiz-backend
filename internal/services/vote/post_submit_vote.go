package services

import (
	"context"
	"fmt"
	"time"

	"genshin-quiz/config"
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/common"
	vote_repo "genshin-quiz/internal/repository/vote"
	"genshin-quiz/internal/webserver/middleware"

	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
)

func PostVote(
	ctx context.Context,
	app *config.App,
	req oapi.PostVotePollRequestObject,
) (oapi.PostVotePollResponseObject, error) {
	userClaims, ok := middleware.GetUserFromContextOnly(ctx)
	if !ok {
		return nil, common.ErrUserNotInContext
	}

	voteInfo, err := vote_repo.GetVoteByUUID(ctx, app.DB, req.Id, nil)
	if err != nil {
		return nil, fmt.Errorf("vote not found: %w", err)
	}

	if err := validateVoteTime(voteInfo.Vote); err != nil {
		return nil, err
	}

	options, err := vote_repo.GetVoteOptions(ctx, app.DB, voteInfo.Vote.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get vote options: %w", err)
	}

	optionVotes, err := validateAndParseVotes(req.Body.Options, *options, voteInfo.Vote)
	if err != nil {
		return nil, err
	}

	if err := checkUserAlreadyVoted(ctx, app.DB, userClaims.UserID, voteInfo.Vote.ID); err != nil {
		return nil, err
	}

	if err := saveUserVotes(ctx, app, userClaims.UserID, voteInfo.Vote.ID, optionVotes); err != nil {
		return nil, err
	}

	return oapi.PostVotePoll200Response{}, nil
}

func validateVoteTime(vote model.Votes) error {
	now := time.Now()
	if vote.ExpiresAt != nil && vote.ExpiresAt.Before(now) {
		return fmt.Errorf("vote has expired")
	}
	if vote.StartAt.After(now) {
		return fmt.Errorf("vote has not started yet")
	}
	return nil
}

func validateAndParseVotes(
	submittedOptions []oapi.PollVote,
	voteOptions []model.VoteOptions,
	vote model.Votes,
) (map[int64]int32, error) {
	optionUUIDToID := make(map[uuid.UUID]int64)
	for _, opt := range voteOptions {
		optionUUIDToID[opt.OptionUUID] = opt.ID
	}

	totalVotes := 0
	optionVotes := make(map[int64]int32)

	for _, optionVote := range submittedOptions {
		optionID, exists := optionUUIDToID[optionVote.OptionId]
		if !exists {
			return nil, fmt.Errorf("invalid option id: %s", optionVote.OptionId)
		}

		totalVotes += optionVote.Votes
		optionVotes[optionID] = int32(optionVote.Votes)

		if vote.VotesPerOption != nil && *vote.VotesPerOption > 0 {
			if optionVote.Votes > int(*vote.VotesPerOption) {
				return nil, fmt.Errorf("votes per option exceeded limit")
			}
		}
	}

	if totalVotes > int(vote.VotesPerUser) {
		return nil, fmt.Errorf("total votes exceeded limit: %d > %d", totalVotes, vote.VotesPerUser)
	}

	return optionVotes, nil
}

func checkUserAlreadyVoted(ctx context.Context, db qrm.DB, userID, voteID int64) error {
	existingVotes, err := vote_repo.GetUserVotedOptions(ctx, db, userID, voteID)
	if err != nil {
		return fmt.Errorf("failed to get user votes: %w", err)
	}
	if len(*existingVotes) > 0 {
		return fmt.Errorf("user has already voted")
	}
	return nil
}

func saveUserVotes(
	ctx context.Context,
	app *config.App,
	userID, voteID int64,
	optionVotes map[int64]int32,
) error {
	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	for optionID, voteCount := range optionVotes {
		userVote := model.UserVotes{
			VoteID:    voteID,
			UserID:    userID,
			OptionID:  optionID,
			VoteCount: voteCount,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := vote_repo.UpsertUserVote(ctx, tx, userVote); err != nil {
			return fmt.Errorf("failed to insert user vote: %w", err)
		}
	}

	if err := vote_repo.RecalculateVoteStats(ctx, tx, voteID); err != nil {
		return fmt.Errorf("failed to update vote stats: %w", err)
	}

	if err := vote_repo.RecalculateVoteOptionStats(ctx, tx, voteID); err != nil {
		return fmt.Errorf("failed to update option stats: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
