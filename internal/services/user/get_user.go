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

	// 处理可选字段，避免空指针解引用
	avatarURL := ""
	if userInfo.AvatarURL != nil {
		avatarURL = *userInfo.AvatarURL
	}

	country := ""
	if userInfo.Country != nil {
		country = *userInfo.Country
	}

	nickname := ""
	if userInfo.DisplayName != nil {
		nickname = *userInfo.DisplayName
	}

	var language *string
	if userInfo.Language != nil {
		language = userInfo.Language
	}

	return &oapi.User{
		Uuid:             userInfo.UserUUID,
		AvatarUrl:        avatarURL,
		Country:          country,
		Language:         language,
		LastLoginAt:      userInfo.CreatedAt,
		Nickname:         nickname,
		RegisteredAt:     userInfo.CreatedAt,
		QuestionsCreated: int(userInfo.QuestionsCreated),
		TotalAnswers:     int(userInfo.TotalSubmissions),
		CorrectAnswers:   int(userInfo.CorrectSubmissions),
		Votes:            int(userInfo.TotalVotes),
	}, nil
}
