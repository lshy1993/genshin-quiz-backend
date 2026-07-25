package services

import (
	"context"

	pg "github.com/go-jet/jet/v2/postgres"

	"genshin-quiz/config"
	"genshin-quiz/generated/db/genshinquiz/public/table"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/common"
	vote_repo "genshin-quiz/internal/repository/vote"
	"genshin-quiz/internal/webserver/middleware"
)

func PostLikePoll(
	ctx context.Context,
	app *config.App,
	req oapi.PostLikePollRequestObject,
) error {
	userClaims, ok := middleware.GetUserFromContextOnly(ctx)
	if !ok {
		return common.ErrUserNotInContext
	}

	voteUUID := req.Id
	value := int16(req.Body.Like)

	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// 获取投票 ID
	voteTbl := table.Votes
	voteIDStmt := pg.SELECT(voteTbl.ID).
		FROM(voteTbl).
		WHERE(voteTbl.VoteUUID.EQ(pg.UUID(voteUUID)))

	var voteIDResult struct {
		ID int64 `alias:"votes.id"`
	}
	err = voteIDStmt.QueryContext(ctx, tx, &voteIDResult)
	if err != nil {
		tx.Rollback()
		return err
	}

	voteID := voteIDResult.ID

	// 更新点赞状态
	err = vote_repo.UpsertVoteLike(ctx, tx, userClaims.UserID, voteUUID, value)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 更新投票的点赞数
	err = vote_repo.UpdateVoteLikeCount(ctx, tx, voteID)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
