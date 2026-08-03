package services

import (
	"context"
	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/dao/transformer"
	user_repo "genshin-quiz/internal/repository/user"
)

func GetUser(
	ctx context.Context,
	app *config.App,
	req oapi.GetUserRequestObject,
) (*oapi.UserPublic, error) {
	// 根据UUID获取用户信息
	userInfo, err := user_repo.GetUserInfoByUUID(ctx, app.DB, req.Id)
	if err != nil {
		return nil, err
	}
	// 内部ID
	userID := userInfo.ID
	userProfile, err := user_repo.GetUserProfileByID(ctx, app.DB, userID)
	if err != nil {
		return nil, err
	}
	userPrivacies, err := user_repo.GetUserPrivaciesByID(ctx, app.DB, userID)
	if err != nil {
		return nil, err
	}
	userStats, err := user_repo.GetUserStatisticsByID(ctx, app.DB, userID)
	if err != nil {
		return nil, err
	}

	res := transformer.UserModelToPublic(*userInfo, *userProfile, *userPrivacies, *userStats)
	return &res, nil
}
