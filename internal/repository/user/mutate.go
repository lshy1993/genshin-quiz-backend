package user_repo

import (
	"context"
	"time"

	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/db/genshinquiz/public/table"

	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func Insert(
	ctx context.Context,
	db qrm.DB,
	email string,
) (*model.Users, error) {
	tbl := table.Users

	newUUID := uuid.New()
	strUUID := newUUID.String()

	insertStmt := tbl.INSERT().MODEL(model.Users{
		UserUUID:    newUUID,
		Email:       email,
		DisplayName: &strUUID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})

	var model model.Users
	err := insertStmt.QueryContext(ctx, db, &model)
	if err != nil {
		return nil, err
	}

	return &model, nil
}

func InsertAuth(
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
	insertStmt := tbl.INSERT().MODEL(model.UserPasswords{
		UserID:       userID,
		PasswordHash: string(hashedPwd),
		CreatedAt:    now,
		UpdatedAt:    now,
	})

	_, err = insertStmt.ExecContext(ctx, db)
	if err != nil {
		return err
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
	model := model.UserLoginLogs{
		UserID:    userID,
		IPAddress: ip,
		LoginAt:   now,
	}

	insertStmt := tbl.INSERT().MODEL(model)
	_, err := insertStmt.ExecContext(ctx, db)
	if err != nil {
		return nil, err
	}
	return &model, nil
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
