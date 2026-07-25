package dao

import (
	"genshin-quiz/generated/db/genshinquiz/public/model"
)

type UserInfoWithAuth struct {
	model.Users
	model.UserCredentials
}
