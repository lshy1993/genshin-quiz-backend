package services

import (
	"context"
	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/common"
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

	avatar := ""
	if userInfo.AvatarURL != nil {
		avatar = *userInfo.AvatarURL
	}
	country := "Unknown"
	if userInfo.Country != nil {
		country = *userInfo.Country
	}
	displayName := ""
	if userInfo.DisplayName != nil {
		displayName = *userInfo.DisplayName
	}

	return &oapi.UserPrivate{
		Uuid:             userInfo.UserUUID,
		AvatarUrl:        avatar,
		Country:          &country,
		Language:         *userInfo.Language,
		LastLoginAt:      userInfo.CreatedAt,
		LastLoginIp:      nil,
		Nickname:         displayName,
		RegisteredAt:     userInfo.CreatedAt,
		QuestionsCreated: int(userInfo.QuestionsCreated),
		TotalAnswers:     int(userInfo.TotalSubmissions),
		CorrectAnswers:   int(userInfo.CorrectSubmissions),
		PollsCreated:     int(userInfo.TotalVotes),
	}, nil
}
