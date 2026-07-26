package dao

import (
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/internal/enum"
	"time"
)

type UserInfoWithAuth struct {
	model.Users
	model.UserCredentials
}
type UpdateUserProfileParams struct {
	Gender   *int16
	Country  *string
	Timezone *string
	Birthday *time.Time
	Website  *string
	Twitter  *string
	Discord  *string
}

type UpdateUserPrivaciesParams struct {
	EmailVisibility    *int16
	BirthdayVisibility *int16
	GenderVisibility   *int16
	CountryVisibility  *int16
}

type UpdateUserParams struct {
	Nickname  *string
	AvatarURL *string
	Language  *string
	Biography *string
}

type LeaderboardParams struct {
	SortBy   enum.LeaderboardSortBy
	SortDesc bool
	Limit    int
	Offset   int
}
type LeaderboardRow struct {
	User    model.Users
	Profile model.UserProfiles
	Privacy model.UserPrivacies
	Stats   model.UserStats
	Total   int
}
