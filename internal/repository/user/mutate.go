package user_repo

import (
	"context"
	"log"
	"strings"
	"time"

	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/db/genshinquiz/public/table"

	"genshin-quiz/internal/common"
	"genshin-quiz/internal/dao"

	"github.com/go-errors/errors"
	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
)

func InsertUser(
	ctx context.Context,
	db qrm.DB,
	email string,
	language string,
) (*model.Users, error) {
	tbl := table.Users

	newUUID := uuid.New()
	strUUID := newUUID.String()
	idx := strings.LastIndex(strUUID, "-")
	tmpName := strUUID
	if idx != -1 && idx+1 < len(strUUID) {
		tmpName = strUUID[idx+1:]
	}
	tmpName = "guest_" + tmpName

	insertStmt := tbl.INSERT(
		tbl.UserUUID,
		tbl.Email,
		tbl.Nickname,
		tbl.Language,
		tbl.CreatedAt,
		tbl.UpdatedAt,
	).
		MODEL(model.Users{
			UserUUID:  newUUID,
			Email:     email,
			Nickname:  tmpName,
			Language:  language,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}).
		RETURNING(tbl.AllColumns)

	var result model.Users
	err := insertStmt.QueryContext(ctx, db, &result)
	if err != nil {
		errStr := err.Error()
		log.Print(errStr)
		if errStr != "" &&
			(contains(errStr, "duplicate key") || contains(errStr, "unique constraint")) {
			return nil, common.ErrUserAlreadyExists
		}
		return nil, errors.WrapPrefix(err, "insert user failed", 0)
	}
	return &result, nil
}

func UpdateUser(
	ctx context.Context,
	db qrm.DB,
	userID int64,
	params dao.UpdateUserParams,
) (*model.Users, error) {
	tbl := table.Users

	columns := pg.ColumnList{tbl.UpdatedAt}
	m := model.Users{UpdatedAt: time.Now()}

	if params.Nickname != nil {
		columns = append(columns, tbl.Nickname)
		m.Nickname = *params.Nickname
	}
	if params.AvatarURL != nil {
		columns = append(columns, tbl.AvatarURL)
		m.AvatarURL = params.AvatarURL
	}
	if params.Language != nil {
		columns = append(columns, tbl.Language)
		m.Language = *params.Language
	}
	if params.Biography != nil {
		columns = append(columns, tbl.Biography)
		m.Biography = params.Biography
	}

	updateStmt := tbl.UPDATE(columns).
		MODEL(m).
		WHERE(tbl.ID.EQ(pg.Int64(userID))).
		RETURNING(tbl.AllColumns)

	var result model.Users
	err := updateStmt.QueryContext(ctx, db, &result)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, common.ErrUserNotFound
		}
		return nil, errors.WrapPrefix(err, "update user failed", 0)
	}
	return &result, nil
}

func InsertUserProfile(
	ctx context.Context,
	db qrm.DB,
	userID int64,
) (*model.UserProfiles, error) {
	tbl := table.UserProfiles
	now := time.Now()

	insertStmt := tbl.INSERT(
		tbl.UserID,
		tbl.Gender,
		tbl.CreatedAt,
		tbl.UpdatedAt,
	).MODEL(model.UserProfiles{
		UserID:    userID,
		Gender:    0, // Unknown
		CreatedAt: now,
		UpdatedAt: now,
	}).RETURNING(tbl.AllColumns)

	var profile model.UserProfiles
	err := insertStmt.QueryContext(ctx, db, &profile)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func UpdateUserProfile(
	ctx context.Context,
	db qrm.DB,
	userID int64,
	params dao.UpdateUserProfileParams,
) (*model.UserProfiles, error) {
	tbl := table.UserProfiles

	columns := pg.ColumnList{tbl.UpdatedAt}
	m := model.UserProfiles{UpdatedAt: time.Now()}

	if params.Gender != nil {
		columns = append(columns, tbl.Gender)
		m.Gender = *params.Gender
	}
	if params.Country != nil {
		columns = append(columns, tbl.Country)
		m.Country = params.Country
	}
	// if params.Timezone != nil {
	// 	columns = append(columns, tbl.Timezone)
	// 	m.Timezone = params.Timezone
	// }
	if params.Birthday != nil {
		columns = append(columns, tbl.Birthday)
		m.Birthday = params.Birthday
	}

	updateStmt := tbl.UPDATE(columns).
		MODEL(m).
		WHERE(tbl.UserID.EQ(pg.Int64(userID))).
		RETURNING(tbl.AllColumns)

	var result model.UserProfiles
	err := updateStmt.QueryContext(ctx, db, &result)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, common.ErrNotFound
		}
		return nil, errors.WrapPrefix(err, "update user profile error", 0)
	}

	return &result, nil
}

func InsertUserPrivacies(
	ctx context.Context,
	db qrm.DB,
	userID int64,
) (*model.UserPrivacies, error) {
	tbl := table.UserPrivacies

	insertStmt := tbl.INSERT(
		tbl.UserID,
		tbl.EmailVisibility,
		tbl.BirthdayVisibility,
		tbl.GenderVisibility,
		tbl.CountryVisibility,
	).MODEL(model.UserPrivacies{
		UserID:             userID,
		EmailVisibility:    0, // private
		BirthdayVisibility: 0, // private
		GenderVisibility:   0, // private
		CountryVisibility:  0, // private
	}).RETURNING(tbl.AllColumns)

	var privacies model.UserPrivacies
	err := insertStmt.QueryContext(ctx, db, &privacies)
	if err != nil {
		return nil, err
	}

	return &privacies, nil
}

func UpdateUserPrivacies(
	ctx context.Context,
	db qrm.DB,
	userID int64,
	params dao.UpdateUserPrivaciesParams,
) (*model.UserPrivacies, error) {
	tbl := table.UserPrivacies

	columns := pg.ColumnList{tbl.UpdatedAt}
	m := model.UserPrivacies{UpdatedAt: time.Now()}

	if params.EmailVisibility != nil {
		columns = append(columns, tbl.EmailVisibility)
		m.EmailVisibility = *params.EmailVisibility
	}
	if params.BirthdayVisibility != nil {
		columns = append(columns, tbl.BirthdayVisibility)
		m.BirthdayVisibility = *params.BirthdayVisibility
	}
	if params.GenderVisibility != nil {
		columns = append(columns, tbl.GenderVisibility)
		m.GenderVisibility = *params.GenderVisibility
	}
	if params.CountryVisibility != nil {
		columns = append(columns, tbl.CountryVisibility)
		m.CountryVisibility = *params.CountryVisibility
	}

	updateStmt := tbl.UPDATE(columns).
		MODEL(m).
		WHERE(tbl.UserID.EQ(pg.Int64(userID))).
		RETURNING(tbl.AllColumns)

	var result model.UserPrivacies
	err := updateStmt.QueryContext(ctx, db, &result)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, common.ErrNotFound
		}
		return nil, errors.WrapPrefix(err, "update user privacies error", 0)
	}

	return &result, nil
}

func DeleteUser(
	ctx context.Context,
	db qrm.DB,
	uuid uuid.UUID,
) error {
	// start := time.Now()
	return nil
}

func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && (index(s, substr) >= 0)
}

// index returns the index of the first instance of substr in s, or -1 if substr is not present in s.
func index(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
