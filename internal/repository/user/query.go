package user_repo

import (
	"context"

	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/db/genshinquiz/public/table"

	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"golang.org/x/crypto/bcrypt"
)

func GetUserByEmailAndPassword(
	ctx context.Context,
	db qrm.DB,
	email string,
	pwd string,
) (*model.Users, error) {
	userTbl := table.Users
	authTbl := table.UserPasswords

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	stmt := pg.SELECT(userTbl.AllColumns).FROM(
		userTbl.LEFT_JOIN(authTbl, authTbl.ID.EQ(userTbl.ID)),
	).WHERE(
		userTbl.Email.EQ(pg.String(email)).AND(
			authTbl.PasswordHash.EQ(pg.String(string(hashedPwd))),
		),
	)

	var user model.Users
	err = stmt.QueryContext(ctx, db, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserInfoByID(
	ctx context.Context,
	db qrm.DB,
	id int64,
) (*model.Users, error) {
	tbl := table.Users
	stmt := pg.SELECT(tbl.AllColumns).FROM(tbl).WHERE(
		tbl.ID.EQ(pg.Int64(id)),
	)

	var user model.Users
	err := stmt.QueryContext(ctx, db, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
