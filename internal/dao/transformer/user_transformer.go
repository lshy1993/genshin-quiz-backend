package transformer

import (
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/logger"

	"github.com/oapi-codegen/runtime/types"
	"go.uber.org/zap"
)

func genderToDTO(g int16) oapi.Gender {
	switch g {
	case 0:
		return oapi.GenderUnknown
	case 1:
		return oapi.GenderMale
	case 2:
		return oapi.GenderFemale
	case 3:
		return oapi.GenderOther
	default:
		// 理论上不该出现，记录一下方便排查脏数据
		logger.L.Warn("unexpected gender value in db", zap.Int16("gender", g))
		return oapi.GenderUnknown
	}
}

func visibilityToDTO(v int16) oapi.Visibility {
	switch v {
	case 0:
		return "private"
	case 1:
		return "public"
	// case 2:
	// 	return oapi.VisibilityFriends
	default:
		logger.L.Warn("unexpected visibility value in db", zap.Int16("visibility", v))
		return "private" // 未知值保守兜底成 private，而不是 public，避免意外泄露
	}
}

func UserModelToPrivate(
	user model.Users,
	profile model.UserProfiles,
	privacies model.UserPrivacies,
	stats model.UserStats,
	login model.UserLoginLogs,
) oapi.UserPrivate {
	nickName := user.Nickname
	avatarURL := ""
	if user.AvatarURL != nil {
		avatarURL = *user.AvatarURL
	}
	country := ""
	if profile.Country != nil {
		country = *profile.Country
	}
	bio := ""
	if user.Biography != nil {
		bio = *user.Biography
	}
	createdIP := ""
	if user.CreatedIP != nil {
		createdIP = *user.CreatedIP
	}

	var genderDTO *oapi.Gender
	g := genderToDTO(profile.Gender)
	genderDTO = &g

	var birthday *types.Date
	if profile.Birthday != nil {
		birthday = &types.Date{Time: *profile.Birthday}
	}

	return oapi.UserPrivate{
		Uuid:         user.UserUUID,
		Nickname:     nickName,
		AvatarUrl:    avatarURL,
		Bio:          bio,
		Birthday:     birthday,
		Country:      &country,
		Gender:       genderDTO,
		Email:        (types.Email)(user.Email),
		Language:     user.Language,
		RegisteredIp: createdIP,
		RegisteredAt: user.CreatedAt,
		LastLoginIp:  &login.IPAddress,
		LastLoginAt:  login.LoginAt,

		BirthdayVisibility: visibilityToDTO(privacies.BirthdayVisibility),
		CountryVisibility:  visibilityToDTO(privacies.CountryVisibility),
		EmailVisibility:    visibilityToDTO(privacies.EmailVisibility),
		GenderVisibility:   visibilityToDTO(privacies.GenderVisibility),

		EmailVerified: user.EmailVerified,

		QuestionsCreated: int(stats.QuestionsCreated),
		TotalAnswers:     int(stats.TotalSubmissions),
		CorrectAnswers:   int(stats.CorrectSubmissions),
		PollsCreated:     int(stats.PollsCreated),
		LikesReceived:    int(stats.LikesReceived),
	}
}

func UserModelToPublic(
	user model.Users,
	profile model.UserProfiles,
	privacies model.UserPrivacies,
	stats model.UserStats,
) oapi.UserPublic {
	avatarURL := ""
	if user.AvatarURL != nil {
		avatarURL = *user.AvatarURL
	}
	country := ""
	if profile.Country != nil && privacies.CountryVisibility == 1 {
		country = *profile.Country
	}
	bio := ""
	if user.Biography != nil {
		bio = *user.Biography
	}

	var genderDTO *oapi.Gender
	if privacies.GenderVisibility == 1 {
		g := genderToDTO(profile.Gender)
		genderDTO = &g
	}

	var birthday *types.Date
	if profile.Birthday != nil && privacies.BirthdayVisibility == 1 {
		birthday = &types.Date{Time: *profile.Birthday}
	}

	return oapi.UserPublic{
		Uuid:             user.UserUUID,
		Nickname:         user.Nickname,
		AvatarUrl:        avatarURL,
		Bio:              bio,
		Birthday:         birthday,
		Country:          &country,
		Gender:           genderDTO,
		Email:            (*types.Email)(&user.Email),
		Language:         user.Language,
		RegisteredAt:     user.CreatedAt,
		QuestionsCreated: int(stats.QuestionsCreated),
		TotalAnswers:     int(stats.TotalSubmissions),
		CorrectAnswers:   int(stats.CorrectSubmissions),
		PollsCreated:     int(stats.PollsCreated),
		LikesReceived:    int(stats.LikesReceived),
	}
}
