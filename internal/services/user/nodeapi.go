package services

import (
	"context"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

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
	Nickname *string      `json:"nickname"`
	Bio      *string      `json:"bio"`
	Language *string      `json:"language"`
	Country  *string      `json:"country"`
	Gender   *oapi.Gender `json:"gender"`
	Birthday *types.Date  `json:"birthday"`

	EmailVisibility    *string `json:"email_visibility"`
	BirthdayVisibility *string `json:"birthday_visibility"`
	GenderVisibility   *string `json:"gender_visibility"`
	CountryVisibility  *string `json:"country_visibility"`
}

func visibilityValue(v *string) (*int16, error) {
	if v == nil {
		return nil, nil
	}
	var value int16
	switch *v {
	case "private":
		value = 0
	case "public":
		value = 1
	default:
		return nil, common.NewBadRequestError("可见性只能是 public 或 private")
	}
	return &value, nil
}

func (r ProfileUpdateRequest) privacyParams() (dao.UpdateUserPrivaciesParams, error) {
	var params dao.UpdateUserPrivaciesParams
	var err error
	if params.EmailVisibility, err = visibilityValue(r.EmailVisibility); err != nil {
		return params, err
	}
	if params.BirthdayVisibility, err = visibilityValue(r.BirthdayVisibility); err != nil {
		return params, err
	}
	if params.GenderVisibility, err = visibilityValue(r.GenderVisibility); err != nil {
		return params, err
	}
	if params.CountryVisibility, err = visibilityValue(r.CountryVisibility); err != nil {
		return params, err
	}
	return params, nil
}

func (r ProfileUpdateRequest) hasPrivacyChange() bool {
	return r.EmailVisibility != nil || r.BirthdayVisibility != nil ||
		r.GenderVisibility != nil || r.CountryVisibility != nil
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

var (
	countryPattern  = regexp.MustCompile(`^[A-Za-z]{2}$`)
	languagePattern = regexp.MustCompile(`^[A-Za-z]{2,3}(-[A-Za-z0-9]{2,5})?$`)
)

func validateProfileRequest(request *ProfileUpdateRequest) error {
	if request.Nickname != nil {
		nickname := strings.TrimSpace(*request.Nickname)
		if nickname == "" || utf8.RuneCountInString(nickname) > 100 {
			return common.NewBadRequestError("昵称长度需为 1-100 个字符")
		}
		request.Nickname = &nickname
	}
	if request.Bio != nil && utf8.RuneCountInString(*request.Bio) > 500 {
		return common.NewBadRequestError("个人简介不能超过 500 个字符")
	}
	if request.Language != nil && !languagePattern.MatchString(*request.Language) {
		return common.NewBadRequestError("语言代码无效")
	}
	if request.Country != nil {
		if !countryPattern.MatchString(*request.Country) {
			return common.NewBadRequestError("国家/地区代码必须是 2 位字母")
		}
		country := strings.ToUpper(*request.Country)
		request.Country = &country
	}
	if request.Birthday != nil {
		if request.Birthday.After(time.Now()) || request.Birthday.Year() < 1900 {
			return common.NewBadRequestError("生日无效")
		}
	}
	return nil
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
		request.Bio == nil &&
		request.Language == nil &&
		request.Country == nil &&
		request.Gender == nil &&
		request.Birthday == nil &&
		!request.hasPrivacyChange() {
		return nil, common.NewBadRequestError("At least one profile field must be provided")
	}
	if err := validateProfileRequest(&request); err != nil {
		return nil, err
	}
	privacyParams, err := request.privacyParams()
	if err != nil {
		return nil, err
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
	if request.hasPrivacyChange() {
		if _, err = user_repo.UpdateUserPrivacies(ctx, tx, userClaims.UserID, privacyParams); err != nil {
			return nil, err
		}
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
