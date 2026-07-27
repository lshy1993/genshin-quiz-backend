package poll_repo

import (
	"context"
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/db/genshinquiz/public/table"

	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
)

func GetPollTransByID(
	ctx context.Context,
	db qrm.DB,
	pollID int64,
) ([]model.PollTranslations, error) {
	tbl := table.PollTranslations

	stmt := pg.SELECT(tbl.AllColumns).
		FROM(tbl).
		WHERE(tbl.PollID.EQ(pg.Int64(pollID)))

	var rows []model.PollTranslations
	err := stmt.QueryContext(ctx, db, &rows)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func GetPollTransByIDs(
	ctx context.Context,
	db qrm.DB,
	pollIDs []int64,
) (map[int64][]model.PollTranslations, error) {
	if len(pollIDs) == 0 {
		return make(map[int64][]model.PollTranslations), nil
	}

	idList := make([]pg.Expression, 0, len(pollIDs))
	for _, id := range pollIDs {
		idList = append(idList, pg.Int64(id))
	}

	tbl := table.PollTranslations

	stmt := pg.SELECT(
		tbl.AllColumns,
	).FROM(
		tbl,
	).WHERE(
		tbl.PollID.IN(idList...),
	)

	var rows []model.PollTranslations
	err := stmt.QueryContext(ctx, db, &rows)
	if err != nil {
		return nil, err
	}

	result := make(map[int64][]model.PollTranslations, len(rows))
	for _, row := range rows {
		result[row.PollID] = append(result[row.PollID], row)
	}

	return result, nil
}
