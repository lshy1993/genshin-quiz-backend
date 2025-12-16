package vote_repo

import (
	"context"

	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/db/genshinquiz/public/table"

	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
)

func InsertVote(
	ctx context.Context,
	db qrm.DB,
	insertModel model.Votes,
) (*model.Votes, error) {
	tbl := table.Votes
	insertStmt := tbl.INSERT(
		tbl.VoteUUID,
		tbl.Public,
		tbl.Category,
		tbl.CreatedBy,
		tbl.CreatedAt,
	).MODEL(insertModel).
		RETURNING(tbl.AllColumns)

	var vote model.Votes
	err := insertStmt.QueryContext(ctx, db, &vote)
	if err != nil {
		return nil, err
	}

	return &vote, nil
}

func InsertVoteTranslations(
	ctx context.Context,
	db qrm.DB,
	translations []model.VoteTranslations,
) error {
	if len(translations) == 0 {
		return nil
	}

	tbl := table.VoteTranslations
	insertStmt := tbl.INSERT(
		tbl.VoteID,
		tbl.Language,
		tbl.Title,
		tbl.Description,
		tbl.CreatedAt,
		tbl.UpdatedAt,
	).MODELS(translations)

	_, err := insertStmt.ExecContext(ctx, db)
	return err
}

func InsertVoteOptions(
	ctx context.Context,
	db qrm.DB,
	options []model.VoteOptions,
) (*[]model.VoteOptions, error) {
	tbl := table.VoteOptions
	insertStmt := tbl.INSERT(
		tbl.VoteID,
		tbl.OptionUUID,
		tbl.CreatedAt,
	).MODELS(options).RETURNING(tbl.AllColumns)

	var insertedOptions []model.VoteOptions
	err := insertStmt.QueryContext(ctx, db, &insertedOptions)
	if err != nil {
		return nil, err
	}

	return &insertedOptions, nil
}

func InsertOptionTranslations(
	ctx context.Context,
	db qrm.DB,
	optionTranslations []model.VoteOptionTranslations,
) error {
	if len(optionTranslations) == 0 {
		return nil
	}

	tbl := table.VoteOptionTranslations
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
	tbl := table.VoteOptions

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
