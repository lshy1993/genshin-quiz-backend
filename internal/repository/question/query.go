package question_repo

import (
	"context"

	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/db/genshinquiz/public/table"

	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
)

func GetQuestionByUUID(ctx context.Context, db qrm.DB, uuid uuid.UUID) (*model.Questions, error) {
	tbl := table.Questions
	stmt := pg.SELECT(tbl.AllColumns).FROM(
		tbl,
	).WHERE(
		tbl.QuestionUUID.EQ(pg.UUID(uuid)),
	)

	var question model.Questions
	err := stmt.QueryContext(ctx, db, &question)
	if err != nil {
		return nil, err
	}

	return &question, nil
}
