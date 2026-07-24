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
	if user.Country != nil {
		country = *user.Country
	}

	return oapi.User{
		Uuid:             user.UserUUID,
		Nickname:         nickName,
		AvatarUrl:        avatarURL,
		Country:          country,
		RegisteredIp:     user.CreatedIP,
		RegisteredAt:     user.CreatedAt,
		LastLoginIp:      &login.IPAddress,
		LastLoginAt:      login.LoginAt,
		QuestionsCreated: 0,
		TotalAnswers:     0,
		CorrectAnswers:   0,
		Votes:            0,
	}
}
