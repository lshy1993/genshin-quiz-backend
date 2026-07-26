package poll_repo

import (
	"context"

	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/db/genshinquiz/public/table"

	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
)

func InsertPoll(
	ctx context.Context,
	db qrm.DB,
	insertModel model.Polls,
) (*model.Polls, error) {
	tbl := table.Polls
	insertStmt := tbl.INSERT(tbl.MutableColumns).
		MODEL(insertModel).
		RETURNING(tbl.AllColumns)

	var poll model.Polls
	err := insertStmt.QueryContext(ctx, db, &poll)
	if err != nil {
		return nil, err
	}

	return &poll, nil
}

func InsertPollTranslations(
	ctx context.Context,
	db qrm.DB,
	translations []model.PollTranslations,
) error {
	if len(translations) == 0 {
		return nil
	}

	tbl := table.PollTranslations
	insertStmt := tbl.INSERT(
		tbl.PollID,
		tbl.Language,
		tbl.Title,
		tbl.Description,
		tbl.CreatedAt,
		tbl.UpdatedAt,
	).MODELS(translations)

	_, err := insertStmt.ExecContext(ctx, db)
	return err
}

func InsertPollOptions(
	ctx context.Context,
	db qrm.DB,
	options []model.PollOptions,
) (*[]model.PollOptions, error) {
	tbl := table.PollOptions
	insertStmt := tbl.INSERT(
		tbl.PollID,
		tbl.OptionUUID,
		tbl.CreatedAt,
	).MODELS(options).RETURNING(tbl.AllColumns)

	var insertedOptions []model.PollOptions
	err := insertStmt.QueryContext(ctx, db, &insertedOptions)
	if err != nil {
		return nil, err
	}

	return &insertedOptions, nil
}

func InsertOptionTranslations(
	ctx context.Context,
	db qrm.DB,
	optionTranslations []model.PollOptionTranslations,
) error {
	if len(optionTranslations) == 0 {
		return nil
	}

	tbl := table.PollOptionTranslations
	insertStmt := tbl.INSERT(
		tbl.OptionID,
		tbl.Language,
		tbl.OptionText,
		tbl.CreatedAt,
		tbl.UpdatedAt,
	).MODELS(optionTranslations)

	_, err := insertStmt.ExecContext(ctx, db)
	return err
}

func UpdateOptionSelected(
	ctx context.Context,
	db qrm.DB,
	optionIDs []int64,
) error {
	tbl := table.PollOptions

	uuidSlice := make([]pg.Expression, 0, len(optionIDs))
	for _, id := range optionIDs {
		uuidSlice = append(uuidSlice, pg.Int64(id))
	}

	updateStmt := tbl.UPDATE().
		SET(
			tbl.VoteCount.SET(tbl.VoteCount.ADD(pg.Int(1))),
		).WHERE(
		tbl.ID.IN(uuidSlice...),
	)
	_, err := updateStmt.ExecContext(ctx, db)
	return err
}

func InsertUserVote(
	ctx context.Context,
	db qrm.DB,
	userVote model.UserVotes,
) error {
	tbl := table.UserVotes
	insertStmt := tbl.INSERT(
		tbl.PollID,
		tbl.UserID,
		tbl.OptionID,
		tbl.VoteCount,
		tbl.CreatedAt,
		tbl.UpdatedAt,
	).MODEL(userVote)

	_, err := insertStmt.ExecContext(ctx, db)
	return err
}

func UpsertUserVote(
	ctx context.Context,
	db qrm.DB,
	userVote model.UserVotes,
) error {
	tbl := table.UserVotes
	insertStmt := tbl.INSERT(
		tbl.PollID,
		tbl.UserID,
		tbl.OptionID,
		tbl.VoteCount,
		tbl.CreatedAt,
		tbl.UpdatedAt,
	).MODEL(userVote).
		ON_CONFLICT(tbl.PollID, tbl.UserID, tbl.OptionID).
		DO_UPDATE(
			pg.SET(
				tbl.VoteCount.SET(pg.Int32(userVote.VoteCount)),
				tbl.UpdatedAt.SET(pg.CURRENT_TIMESTAMP()),
			),
		)

	_, err := insertStmt.ExecContext(ctx, db)
	return err
}
