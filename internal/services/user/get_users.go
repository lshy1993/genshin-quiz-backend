package services

import (
	"context"
	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	user_repo "genshin-quiz/internal/repository/user"
)

func GetUsers(
	ctx context.Context,
	app *config.App,
	req oapi.GetUsersRequestObject,
) (*oapi.GetUsers200JSONResponse, error) {
	// 获取用户信息
	userInfos, err := user_repo.GetUserInfosByUUIDs(ctx, app.DB, *req.Params.Ids)
	if err != nil {
		return nil, err
	}

	// Convert []*model.Users to []oapi.User
	users := make([]oapi.User, 0, len(userInfos))
	for _, userInfo := range userInfos {
		users = append(users, oapi.User{
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
		})
	}

	return &oapi.GetUsers200JSONResponse{
		Total: len(users),
		Users: users,
	}, nil
}
