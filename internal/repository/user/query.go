package user_repo

import (
	"context"
	"errors"

	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/db/genshinquiz/public/table"

	"genshin-quiz/internal/common"
	"genshin-quiz/internal/dao"

	go_errors "github.com/go-errors/errors"
	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
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
		return nil, common.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user[0], nil
}

func GetPasswordByEmail(
	ctx context.Context,
	db qrm.DB,
	email string,
) (*dao.UserInfoWithAuth, error) {
	tbl := table.Users
	authTbl := table.UserPasswords
	stmt := pg.SELECT(tbl.AllColumns, authTbl.PasswordHash).
		FROM(tbl.LEFT_JOIN(authTbl, tbl.ID.EQ(authTbl.UserID))).
		WHERE(
			tbl.Email.EQ(pg.String(email)),
		)
	var auth []dao.UserInfoWithAuth
	err := stmt.QueryContext(ctx, db, &auth)
	if err != nil {
		return nil, go_errors.WrapPrefix(err, "checking password failed", 0)
	}
	if len(auth) == 0 {
		return nil, common.ErrUserNotFound
	}

	return &auth[0], nil
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
		// 如果是记录不存在的错误，返回标准错误
		if err.Error() == "qrm: no rows in result set" {
			return nil, common.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// CheckUserExists 检查用户是否存在（复用 GetUserInfoByID）
func CheckUserExists(
	ctx context.Context,
	db qrm.DB,
	userID int64,
) (bool, error) {
	_, err := GetUserInfoByID(ctx, db, userID)
	if err != nil {
		if errors.Is(err, common.ErrUserNotFound) {
			return false, nil
		}
		// 其他错误返回
		return false, err
	}

	return true, nil
}
