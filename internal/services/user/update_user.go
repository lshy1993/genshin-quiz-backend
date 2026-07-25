package services

import (
	"context"
	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/common"
	user_repo "genshin-quiz/internal/repository/user"
	"genshin-quiz/internal/webserver/middleware"
)

func UpdateUser(
	ctx context.Context,
	app *config.App,
	req oapi.UpdateUserRequestObject,
) (*oapi.User, error) {
	userClaims, ok := middleware.GetUserFromContextOnly(ctx)
	if !ok {
		return nil, common.ErrUserNotInContext
	}

	// 获取目标用户信息
	existing, err := user_repo.GetUserInfoByUUID(ctx, app.DB, req.Id)
	if err != nil {
		return nil, err
	}

	// 只允许用户修改自己的资料
	if existing.ID != userClaims.UserID {
		return nil, common.ErrUserAuthError
	}

	// 应用可更新字段
	if req.Body != nil {
		nickname := req.Body.Nickname
		existing.DisplayName = &nickname

		avatarURL := req.Body.AvatarUrl
		existing.AvatarURL = &avatarURL

		country := req.Body.Country
		existing.Country = &country

		language := req.Body.Language
		existing.Language = language

		gender := req.Body.Gender
		existing.Gender = (*string)(gender)

		bio := req.Body.Bio
		existing.Biography = bio
	}

	updated, err := user_repo.Update(ctx, app.DB, *existing)
	if err != nil {
		return nil, err
	}

	displayName := ""
	if updated.DisplayName != nil {
		displayName = *updated.DisplayName
	}
	avatarURL := ""
	if updated.AvatarURL != nil {
		avatarURL = *updated.AvatarURL
	}
	country := ""
	if updated.Country != nil {
		country = *updated.Country
	}

	emailVerified := false

	return &oapi.User{
		Uuid:             updated.UserUUID,
		AvatarUrl:        avatarURL,
		Country:          country,
		Language:         updated.Language,
		Gender:           (*oapi.UserGender)(updated.Gender),
		Bio:              updated.Biography,
		EmailVerified:    &emailVerified,
		EmailPublic:      &updated.ShowEmail,
		LastLoginAt:      updated.UpdatedAt,
		Nickname:         displayName,
		RegisteredAt:     updated.CreatedAt,
		QuestionsCreated: int(updated.QuestionsCreated),
		TotalAnswers:     int(updated.TotalSubmissions),
		CorrectAnswers:   int(updated.CorrectSubmissions),
		Votes:            int(updated.TotalVotes),
	}, nil
}
