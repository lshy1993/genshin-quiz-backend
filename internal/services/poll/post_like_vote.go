package services

import (
	"context"

	pg "github.com/go-jet/jet/v2/postgres"

	"genshin-quiz/config"
	"genshin-quiz/generated/db/genshinquiz/public/table"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/common"
	poll_repo "genshin-quiz/internal/repository/poll"
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

	pollUUID := req.Id
	value := int16(req.Body.Like)

	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 获取投票 ID
	pollTbl := table.Polls
	pollIDStmt := pg.SELECT(pollTbl.ID).
		FROM(pollTbl).
		WHERE(pollTbl.PollUUID.EQ(pg.UUID(pollUUID)))

	var pollIDResult struct {
		ID int64 `alias:"polls.id"`
	}
	err = pollIDStmt.QueryContext(ctx, tx, &pollIDResult)
	if err != nil {
		return err
	}

	pollID := pollIDResult.ID

	// 更新点赞状态
	err = poll_repo.UpsertPollLike(ctx, tx, userClaims.UserID, pollUUID, value)
	if err != nil {
		return err
	}

	// 更新投票的点赞数
	err = poll_repo.UpdatePollLikeCount(ctx, tx, pollID)
	if err != nil {
		return err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
