package user_repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/db/genshinquiz/public/table"

	"genshin-quiz/internal/common"

	"github.com/go-errors/errors"
	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
)

func InsertUser(
	ctx context.Context,
	db qrm.DB,
	email string,
	language *string,
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
		tbl.DisplayName,
		tbl.Language,
		tbl.CreatedAt,
		tbl.UpdatedAt,
	).
		MODEL(model.Users{
			UserUUID:    newUUID,
			Email:       email,
			DisplayName: &tmpName,
			Language:    language,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}).
		RETURNING(tbl.AllColumns)

	var result model.Users
	err := insertStmt.QueryContext(ctx, db, &result)
	if err != nil {
		errStr := err.Error()
		fmt.Print(errStr)
		if errStr != "" &&
			(contains(errStr, "duplicate key") || contains(errStr, "unique constraint")) {
			return nil, common.ErrUserAlreadyExists
		}
		return nil, errors.WrapPrefix(err, "insert user failed", 0)
	}
	return &result, nil
}

func Update(
	ctx context.Context,
	db qrm.DB,
	u model.Users,
) (*model.Users, error) {
	tbl := table.Users

	u.UpdatedAt = time.Now()

	updateStmt := tbl.UPDATE(
		tbl.DisplayName,
		tbl.AvatarURL,
		tbl.Country,
		tbl.Language,
		tbl.UpdatedAt,
	).
		MODEL(u).
		WHERE(tbl.ID.EQ(pg.Int64(u.ID))).
		RETURNING(tbl.AllColumns)

	var result model.Users
	err := updateStmt.QueryContext(ctx, db, &result)
	if err != nil {
		return nil, errors.WrapPrefix(err, "update user failed", 0)
	}
	return &result, nil
}

func Delete(
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
