package dao

import (
	"genshin-quiz/generated/db/genshinquiz/public/model"
)

type UserInfoWithAuth struct {
	User model.Users
	Auth model.UserPasswords
}
