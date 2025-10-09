package user_repo

import (
	"context"

	"genshin-quiz/generated/db/genshinquiz/public/model"

	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
)

func Insert(
	ctx context.Context,
	db qrm.DB,
	model model.Users,
) (*model.Users, error) {
	// start := time.Now()
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
