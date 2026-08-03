package services

import (
	"context"
	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/common"
	"genshin-quiz/internal/dao/transformer"
	user_repo "genshin-quiz/internal/repository/user"
	"genshin-quiz/internal/webserver/middleware"
)

func GetMe(
	ctx context.Context,
	app *config.App,
	req oapi.GetCurrentUserRequestObject,
) (*oapi.UserPrivate, error) {
	userClaims, ok := middleware.GetUserFromContextOnly(ctx)
	if !ok {
		return nil, common.ErrUserNotInContext
	}

	// 获取用户信息
	userInfo, err := user_repo.GetUserInfoByID(ctx, app.DB, userClaims.UserID)
	if err != nil {
		return nil, err
	}
	userProfile, err := user_repo.GetUserProfileByID(ctx, app.DB, userClaims.UserID)
	if err != nil {
		return nil, err
	}
	userPrivacies, err := user_repo.GetUserPrivaciesByID(ctx, app.DB, userClaims.UserID)
	if err != nil {
		return nil, err
	}
	userStats, err := user_repo.GetUserStatisticsByID(ctx, app.DB, userClaims.UserID)
	if err != nil {
		return nil, err
	}
	loginInfo, err := user_repo.GetLatestLoginLogByID(ctx, app.DB, userClaims.UserID)
	if err != nil {
		return nil, err
	}

	res := transformer.UserModelToPrivate(
		*userInfo,
		*userProfile,
		*userPrivacies,
		*userStats,
		*loginInfo,
	)

	return &res, nil
}
