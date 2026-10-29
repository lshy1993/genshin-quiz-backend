package services

import (
	"context"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/common"
	"genshin-quiz/internal/dao"
	"genshin-quiz/internal/dao/transformer"
	user_repo "genshin-quiz/internal/repository/user"
	"genshin-quiz/internal/webserver/middleware"

	"github.com/go-errors/errors"
	"github.com/oapi-codegen/runtime/types"
)

type ProfileUpdateRequest struct {
	Nickname  *string      `json:"nickname"`
	AvatarURL *string      `json:"avatar_url"`
	Bio       *string      `json:"bio"`
	Language  *string      `json:"language"`
	Country   *string      `json:"country"`
	Gender    *oapi.Gender `json:"gender"`
	Birthday  *types.Date  `json:"birthday"`
}

func IsEmailAvailable(ctx context.Context, app *config.App, email string) (bool, error) {
	_, err := user_repo.GetUserByEmail(ctx, app.DB, email)
	if err != nil {
		if errors.Is(err, common.ErrUserNotFound) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

func UpdateProfile(
	ctx context.Context,
	app *config.App,
	request ProfileUpdateRequest,
) (*oapi.UserPrivate, error) {
	userClaims, ok := middleware.GetUserFromContextOnly(ctx)
	if !ok {
		return nil, common.ErrUserNotInContext
	}
	if request.Nickname == nil &&
		request.AvatarURL == nil &&
		request.Bio == nil &&
		request.Language == nil &&
		request.Country == nil &&
		request.Gender == nil &&
		request.Birthday == nil {
		return nil, common.NewBadRequestError("At least one profile field must be provided")
	}
	if request.Gender != nil {
		switch *request.Gender {
		case oapi.GenderUnknown, oapi.GenderMale, oapi.GenderFemale, oapi.GenderOther:
		default:
			return nil, common.NewBadRequestError("Invalid gender")
		}
	}

	userParams := dao.UpdateUserParams{
		Nickname:  request.Nickname,
		AvatarURL: request.AvatarURL,
		Language:  request.Language,
		Biography: request.Bio,
	}
	profileParams := dao.UpdateUserProfileParams{
		Country: request.Country,
	}
	if request.Gender != nil {
		gender := DTOToGender(*request.Gender)
		profileParams.Gender = &gender
	}
	if request.Birthday != nil {
		birthday := request.Birthday.Time
		profileParams.Birthday = &birthday
	}

	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.WrapPrefix(err, "failed to begin profile update transaction", 0)
	}
	defer tx.Rollback()

	updatedUser, err := user_repo.UpdateUser(ctx, tx, userClaims.UserID, userParams)
	if err != nil {
		return nil, err
	}
	updatedProfile, err := user_repo.UpdateUserProfile(ctx, tx, userClaims.UserID, profileParams)
	if err != nil {
		return nil, err
	}
	privacy, err := user_repo.GetUserPrivaciesByID(ctx, tx, userClaims.UserID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, errors.WrapPrefix(err, "failed to commit profile update transaction", 0)
	}

	stats, err := user_repo.GetUserStatisticsByID(ctx, app.DB, userClaims.UserID)
	if err != nil {
		return nil, err
	}
	login, err := user_repo.GetLatestLoginLogByID(ctx, app.DB, userClaims.UserID)
	if err != nil {
		return nil, err
	}

	result := transformer.UserModelToPrivate(
		*updatedUser,
		*updatedProfile,
		*privacy,
		*stats,
		*login,
	)
	return &result, nil
}
