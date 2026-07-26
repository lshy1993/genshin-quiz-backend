package services

import (
	"context"
	"fmt"
	"time"

	"genshin-quiz/config"
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/common"
	poll_repo "genshin-quiz/internal/repository/poll"
	"genshin-quiz/internal/webserver/middleware"

	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
)

// 用户投票了.
func PostVotePoll(
	ctx context.Context,
	app *config.App,
	req oapi.PostVotePollRequestObject,
) (oapi.PostVotePollResponseObject, error) {
	userClaims, ok := middleware.GetUserFromContextOnly(ctx)
	if !ok {
		return nil, common.ErrUserNotInContext
	}

	pollInfo, err := poll_repo.GetPollByUUID(ctx, app.DB, req.Id, nil)
	if err != nil {
		return nil, fmt.Errorf("vote not found: %w", err)
	}

	if err := validatePollTime(pollInfo.Poll); err != nil {
		return nil, err
	}

	pollID := pollInfo.Poll.ID
	options, err := poll_repo.GetPollOptions(ctx, app.DB, pollID)
	if err != nil {
		return nil, fmt.Errorf("failed to get vote options: %w", err)
	}

	optionVotes, err := validateAndParseVotes(req.Body.Options, *options, pollInfo.Poll)
	if err != nil {
		return nil, err
	}

	if err := checkUserAlreadyVoted(ctx, app.DB, userClaims.UserID, pollID); err != nil {
		return nil, err
	}

	if err := saveUserVotes(ctx, app, userClaims.UserID, pollID, optionVotes); err != nil {
		return nil, err
	}

	return oapi.PostVotePoll200Response{}, nil
}

func validatePollTime(poll model.Polls) error {
	now := time.Now()
	if poll.ExpiresAt != nil && poll.ExpiresAt.Before(now) {
		return fmt.Errorf("poll has expired")
	}
	if poll.StartAt.After(now) {
		return fmt.Errorf("poll has not started yet")
	}
	return nil
}

func validateAndParseVotes(
	submittedOptions []oapi.PollVote,
	pollOptions []model.PollOptions,
	poll model.Polls,
) (map[int64]int32, error) {
	optionUUIDToID := make(map[uuid.UUID]int64)
	for _, opt := range pollOptions {
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

		if poll.VotesPerOption > 0 {
			if optionVote.Votes > int(poll.VotesPerOption) {
				return nil, fmt.Errorf("votes per option exceeded limit")
			}
		}
	}

	if totalVotes > int(poll.VotesPerUser) {
		return nil, fmt.Errorf("total votes exceeded limit: %d > %d", totalVotes, poll.VotesPerUser)
	}

	return optionVotes, nil
}

func checkUserAlreadyVoted(ctx context.Context, db qrm.DB, userID, pollID int64) error {
	existingVotes, err := poll_repo.GetUserVotedOptions(ctx, db, userID, pollID)
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
			PollID:    voteID,
			UserID:    userID,
			OptionID:  optionID,
			VoteCount: voteCount,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := poll_repo.UpsertUserVote(ctx, tx, userVote); err != nil {
			return fmt.Errorf("failed to insert user vote: %w", err)
		}
	}

	if err := poll_repo.RecalculatePollStats(ctx, tx, voteID); err != nil {
		return fmt.Errorf("failed to update vote stats: %w", err)
	}

	if err := poll_repo.RecalculatePollOptionStats(ctx, tx, voteID); err != nil {
		return fmt.Errorf("failed to update option stats: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
