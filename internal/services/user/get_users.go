package services

import (
	"context"
	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/dao"
	"genshin-quiz/internal/dao/transformer"
	"genshin-quiz/internal/enum"
	user_repo "genshin-quiz/internal/repository/user"
)

func GetUsers(
	ctx context.Context,
	app *config.App,
	req oapi.GetUsersRequestObject,
) (*oapi.GetUsers200JSONResponse, error) {
	limit := 10
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}

	offset := 0
	if req.Params.Offset != nil {
		offset = *req.Params.Offset
	}

	sortBy := enum.SortByAccuracy
	if req.Params.SortBy != nil {
		sortBy = enum.LeaderboardSortBy(*req.Params.SortBy)
	}

	sortDesc := true
	if req.Params.SortDesc != nil {
		sortDesc = *req.Params.SortDesc
	}

	rows, total, err := user_repo.GetUsersLeaderboard(ctx, app.DB, dao.LeaderboardParams{
		SortBy:   sortBy,
		SortDesc: sortDesc,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, err
	}

	users := make([]oapi.UserPublic, 0, len(rows))
	for _, row := range rows {
		users = append(users, transformer.UserModelToPublic(
			row.User, row.Profile, row.Privacy, row.Stats,
		))
	}

	return &oapi.GetUsers200JSONResponse{
		Total: total,
		Users: users,
	}, nil
}
