package transformer

import (
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/oapi"
)

func UserModelToDTO(
	user model.Users,
	login model.UserLoginLogs,
) oapi.User {
	nickName := ""
	if user.DisplayName != nil {
		nickName = *user.DisplayName
	}
	avatarURL := ""
	if user.AvatarURL != nil {
		avatarURL = *user.AvatarURL
	}
	country := ""
	if user.Location != nil {
		country = *user.Location
	}

	return oapi.User{
		Uuid:             user.UserUUID,
		Nickname:         nickName,
		AvatarUrl:        avatarURL,
		Country:          country,
		RegisteredAt:     user.CreatedAt,
		Ip:               login.IPAddress,
		LastLoginAt:      login.LoginAt,
		QuestionsCreated: 0,
		TotalAnswers:     0,
		CorrectAnswers:   0,
		Votes:            0,
	}
}
