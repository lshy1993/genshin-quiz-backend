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
	limit := 10
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}

	offset := 0
	if req.Params.Offset != nil {
		offset = *req.Params.Offset
	}

	sortBy := "accuracy"
	if req.Params.SortBy != nil {
		sortBy = string(*req.Params.SortBy)
	}

	sortDesc := true
	if req.Params.SortDesc != nil {
		sortDesc = *req.Params.SortDesc
	}

	result, err := user_repo.GetUsersLeaderboard(
		ctx,
		app.DB,
		req.Params.Ids,
		limit,
		offset,
		sortBy,
		sortDesc,
	)
	if err != nil {
		return nil, err
	}

	users := make([]oapi.UserPublic, 0, len(result.Users))
	for _, row := range result.Users {
		userInfo := row.User
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
		likesReceived := int(row.LikesReceived)

		users = append(users, oapi.UserPublic{
			Uuid:             userInfo.UserUUID,
			AvatarUrl:        avatarURL,
			Country:          &country,
			Language:         *userInfo.Language,
			Nickname:         nickname,
			RegisteredAt:     userInfo.CreatedAt,
			QuestionsCreated: int(userInfo.QuestionsCreated),
			TotalAnswers:     int(userInfo.TotalSubmissions),
			CorrectAnswers:   int(userInfo.CorrectSubmissions),
			LikesReceived:    likesReceived,
			PollsCreated:     int(userInfo.TotalVotes),
		})
	}

	return &oapi.GetUsers200JSONResponse{
		Total: result.Total,
		Users: users,
	}, nil
}
