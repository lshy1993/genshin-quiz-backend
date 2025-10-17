package user_repo

import (
	"context"

	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/db/genshinquiz/public/table"

	api_error "genshin-quiz/internal/errors"

	"github.com/go-errors/errors"
	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"golang.org/x/crypto/bcrypt"
)

func GetUserByEmail(
	ctx context.Context,
	db qrm.DB,
	email string,
) (*model.Users, error) {
	userTbl := table.Users

	stmt := pg.SELECT(userTbl.AllColumns).
		FROM(userTbl).
		WHERE(
			userTbl.Email.EQ(pg.String(email)),
		)

	var user []model.Users
	err := stmt.QueryContext(ctx, db, &user)
	if len(user) == 0 {
		return nil, api_error.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user[0], nil
}

func CheckPassword(
	ctx context.Context,
	db qrm.DB,
	userId int64,
	pwd string,
) error {
	authTbl := table.UserPasswords

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	stmt := pg.SELECT(authTbl.AllColumns).FROM(authTbl).WHERE(
		authTbl.UserID.EQ(pg.Int64(userId)).AND(
			authTbl.PasswordHash.EQ(pg.String(string(hashedPwd))),
		),
	)
	var user []model.Users
	err = stmt.QueryContext(ctx, db, &user)
	if len(user) == 0 {
		return api_error.ErrInvalidCredentials
	}
	if err != nil {
		return errors.WrapPrefix(err, "checking password failed", 0)
	}

	return nil
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
