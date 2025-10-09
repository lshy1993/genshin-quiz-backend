package user_repo

import (
	"context"

	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/db/genshinquiz/public/table"

	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
)

func GetByID(ctx context.Context, db qrm.DB, id int64) (*model.Users, error) {
	tbl := table.Users
	stmt := pg.SELECT(tbl.AllColumns).FROM(
		tbl,
	).WHERE(
		tbl.ID.EQ(pg.Int64(id)),
	)

	var user model.Users
	err := stmt.QueryContext(ctx, db, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
