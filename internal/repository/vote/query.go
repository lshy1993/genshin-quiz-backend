package vote_repo

import (
	"context"

	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/db/genshinquiz/public/table"

	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
)

func GetVoteByUUID(ctx context.Context, db qrm.DB, uuid uuid.UUID) (*model.Votes, error) {
	tbl := table.Votes
	stmt := pg.SELECT(tbl.AllColumns).FROM(
		tbl,
	).WHERE(
		tbl.VoteUUID.EQ(pg.UUID(uuid)),
	)

	var vote model.Votes
	err := stmt.QueryContext(ctx, db, &vote)
	if err != nil {
		return nil, err
	}

	return &vote, nil
}
