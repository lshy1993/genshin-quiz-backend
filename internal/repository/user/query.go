package user_repo

import (
	"context"

	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/db/genshinquiz/public/table"

	"genshin-quiz/internal/common"

	"github.com/go-errors/errors"
	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
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
		return nil, errors.WrapPrefix(err, "get user base by email failed", 0)
	}

	return &user[0], nil
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
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, common.ErrUserNotFound // 如果是记录不存在的错误，返回标准错误
		}
		return nil, errors.WrapPrefix(err, "get user base by id failed", 0)
	}

	return &user, nil
}

func GetUserProfileByID(
	ctx context.Context,
	db qrm.DB,
	id int64,
) (*model.UserProfiles, error) {
	tbl := table.UserProfiles
	stmt := pg.SELECT(tbl.AllColumns).FROM(tbl).WHERE(
		tbl.UserID.EQ(pg.Int64(id)),
	)

	var profile model.UserProfiles
	err := stmt.QueryContext(ctx, db, &profile)
	if err != nil {
		return nil, errors.WrapPrefix(err, "get user profile failed", 0)
	}

	return &profile, nil
}

func GetUserPrivaciesByID(
	ctx context.Context,
	db qrm.DB,
	id int64,
) (*model.UserPrivacies, error) {
	tbl := table.UserPrivacies
	stmt := pg.SELECT(tbl.AllColumns).FROM(tbl).WHERE(
		tbl.UserID.EQ(pg.Int64(id)),
	)

	var privacies model.UserPrivacies
	err := stmt.QueryContext(ctx, db, &privacies)
	if err != nil {
		return nil, errors.WrapPrefix(err, "get user Privacies failed", 0)
	}

	return &privacies, nil
}

func GetUserStatisticsByID(
	ctx context.Context,
	db qrm.DB,
	id int64,
) (*model.UserStats, error) {
	tbl := table.UserStats
	stmt := pg.SELECT(tbl.AllColumns).FROM(tbl).WHERE(
		tbl.UserID.EQ(pg.Int64(id)),
	)

	var stats model.UserStats
	err := stmt.QueryContext(ctx, db, &stats)
	if err != nil {
		return nil, errors.WrapPrefix(err, "get user Statistics failed", 0)
	}

	return &stats, nil
}

func GetUserInfoByUUID(
	ctx context.Context,
	db qrm.DB,
	uuid uuid.UUID,
) (*model.Users, error) {
	tbl := table.Users
	stmt := pg.SELECT(tbl.AllColumns).FROM(tbl).WHERE(
		tbl.UserUUID.EQ(pg.UUID(uuid)),
	)

	var users []*model.Users
	err := stmt.QueryContext(ctx, db, &users)
	if err != nil {
		return nil, errors.WrapPrefix(err, "get user base failed", 0)
	}
	if len(users) == 0 {
		return nil, common.ErrUserNotFound
	}

	return users[0], nil
}

func GetUserInfosByUUIDs(
	ctx context.Context,
	db qrm.DB,
	uuids []uuid.UUID,
) ([]*model.Users, error) {
	tbl := table.Users

	uuidExp := make([]pg.Expression, len(uuids))
	for i, id := range uuids {
		uuidExp[i] = pg.UUID(id)
	}
	stmt := pg.SELECT(tbl.AllColumns).FROM(tbl).WHERE(
		tbl.UserUUID.IN(uuidExp...),
	)

	var users []*model.Users
	err := stmt.QueryContext(ctx, db, &users)
	if err != nil {
		return nil, errors.WrapPrefix(err, "get user info by uuids failed", 0)
	}

	return users, nil
}

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
		return false, errors.WrapPrefix(err, "get user exists failed", 0)
	}

	return true, nil
}
