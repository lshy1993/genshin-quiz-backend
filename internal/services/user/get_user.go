package services

import (
	"context"
	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	user_repo "genshin-quiz/internal/repository/user"
)

func GetUser(
	ctx context.Context,
	app *config.App,
	req oapi.GetUserRequestObject,
) (*oapi.User, error) {
	// 根据UUID获取用户信息
	userInfo, err := user_repo.GetUserInfoByUUID(ctx, app.DB, req.Id)
	if err != nil {
		return nil, err
	}

	return &oapi.User{
		Uuid:             userInfo.UserUUID,
		AvatarUrl:        *userInfo.AvatarURL,
		Country:          *userInfo.Location,
		Ip:               "",
		Language:         userInfo.Language,
		LastLoginAt:      userInfo.CreatedAt,
		Nickname:         *userInfo.DisplayName,
		RegisteredAt:     userInfo.CreatedAt,
		QuestionsCreated: 0,
		TotalAnswers:     0,
		CorrectAnswers:   0,
		Votes:            0,
	}, nil
}
