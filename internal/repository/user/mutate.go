package user_repo

import (
	"context"
	"fmt"
	"time"

	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/db/genshinquiz/public/table"

	api_error "genshin-quiz/internal/errors"

	"github.com/go-errors/errors"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func InsertUser(
	ctx context.Context,
	db qrm.DB,
	email string,
) (*model.Users, error) {
	tbl := table.Users

	newUUID := uuid.New()
	strUUID := newUUID.String()

	insertStmt := tbl.INSERT(
		tbl.UserUUID,
		tbl.Email,
		tbl.DisplayName,
		tbl.CreatedAt,
		tbl.UpdatedAt,
	).
		MODEL(model.Users{
			UserUUID:    newUUID,
			Email:       email,
			DisplayName: &strUUID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}).
		RETURNING(tbl.AllColumns)

	var result model.Users
	err := insertStmt.QueryContext(ctx, db, &result)
	if err != nil {
		errStr := err.Error()
		fmt.Print(errStr)
		if errStr != "" && (contains(errStr, "duplicate key") || contains(errStr, "unique constraint")) {
			return nil, api_error.ErrUserAlreadyExists
		}
		return nil, errors.WrapPrefix(err, "insert user failed", 0)
	}
	return &result, nil
}

func InsertUserAuth(
	ctx context.Context,
	db qrm.DB,
	userID int64,
	pwd string,
) error {
	tbl := table.UserPasswords

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now()
	insertStmt := tbl.INSERT(
		tbl.UserID,
		tbl.PasswordHash,
		tbl.CreatedAt,
		tbl.UpdatedAt,
	).
		MODEL(model.UserPasswords{
			UserID:       userID,
			PasswordHash: string(hashedPwd),
			CreatedAt:    now,
			UpdatedAt:    now,
		})

	_, err = insertStmt.ExecContext(ctx, db)
	if err != nil {
		errStr := err.Error()
		if errStr != "" && (contains(errStr, "duplicate key") || contains(errStr, "unique constraint")) {
			return api_error.NewBadRequestError("user password already exists")
		}
		return errors.WrapPrefix(err, "insert password failed", 0)
	}

	return nil
}

func InsertLoginLog(
	ctx context.Context,
	db qrm.DB,
	userID int64,
	ip string,
) (*model.UserLoginLogs, error) {
	tbl := table.UserLoginLogs

	now := time.Now()
	insertStmt := tbl.INSERT(
		tbl.UserID,
		tbl.IPAddress,
		tbl.LoginAt,
	).
		MODEL(model.UserLoginLogs{
			UserID:    userID,
			IPAddress: ip,
			LoginAt:   now,
		}).
		RETURNING(tbl.AllColumns)

	var result model.UserLoginLogs
	err := insertStmt.QueryContext(ctx, db, &result)
	if err != nil {
		return nil, errors.WrapPrefix(err, "insert login logs error", 0)
	}
	return &result, nil
}

func Update(
	ctx context.Context,
	db qrm.DB,
	model model.Users,
) (*model.Users, error) {
	// start := time.Now()

	// updateStmt := table.Users.UPDATE()

	return &model, nil
}

func Delete(
	ctx context.Context,
	db qrm.DB,
	uuid uuid.UUID,
) error {
	// start := time.Now()
	return nil
}

// contains 用于兼容 go1.17 及更早版本
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
