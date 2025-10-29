package services

import (
	"context"
	"fmt"
	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	user_repo "genshin-quiz/internal/repository/user"
	"genshin-quiz/internal/webserver/middleware"
)

func GetMe(
	ctx context.Context,
	app *config.App,
	req oapi.GetCurrentUserRequestObject,
) (*oapi.User, error) {
	userClaims, ok := middleware.GetUserFromContextOnly(ctx)
	if !ok {
		return nil, fmt.Errorf("user not found in context")
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
	if userInfo.Location != nil {
		country = *userInfo.Location
	}
	displayName := ""
	if userInfo.DisplayName != nil {
		displayName = *userInfo.DisplayName
	}

	return &oapi.User{
		Uuid:             userInfo.UserUUID,
		AvatarUrl:        avatar,
		Country:          country,
		Ip:               "",
		Language:         userInfo.Language,
		LastLoginAt:      userInfo.CreatedAt,
		Nickname:         displayName,
		RegisteredAt:     userInfo.CreatedAt,
		QuestionsCreated: 0,
		TotalAnswers:     0,
		CorrectAnswers:   0,
		Votes:            0,
	}, nil
}
