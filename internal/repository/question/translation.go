package question_repo

import (
	"context"
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/db/genshinquiz/public/table"

	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
)

func GetQuestionTransByID(
	ctx context.Context,
	db qrm.DB,
	questionID int64,
) ([]model.QuestionTranslations, error) {
	tbl := table.QuestionTranslations

	stmt := pg.SELECT(tbl.AllColumns).
		FROM(tbl).
		WHERE(tbl.QuestionID.EQ(pg.Int64(questionID)))

	var rows []model.QuestionTranslations
	err := stmt.QueryContext(ctx, db, &rows)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func GetQuestionTransByIDs(
	ctx context.Context,
	db qrm.DB,
	questionIDs []int64,
) (map[int64][]model.QuestionTranslations, error) {
	if len(questionIDs) == 0 {
		return make(map[int64][]model.QuestionTranslations), nil
	}

	transTbl := table.QuestionTranslations

	idList := make([]pg.Expression, 0, len(questionIDs))
	for _, id := range questionIDs {
		idList = append(idList, pg.Int64(id))
	}

	stmt := pg.SELECT(
		transTbl.AllColumns,
	).FROM(
		transTbl,
	).WHERE(
		transTbl.QuestionID.IN(idList...),
	)

	var rows []model.QuestionTranslations
	err := stmt.QueryContext(ctx, db, &rows)
	if err != nil {
		return nil, err
	}

	result := make(map[int64][]model.QuestionTranslations, len(questionIDs))
	for _, row := range rows {
		result[row.QuestionID] = append(result[row.QuestionID], row)
	}

	return result, nil
}
