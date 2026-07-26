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
	"genshin-quiz/logger"

	"github.com/go-errors/errors"
	"go.uber.org/zap"
)

func DTOToGender(g oapi.Gender) int16 {
	switch g {
	case oapi.GenderUnknown:
		return 0
	case oapi.GenderMale:
		return 1
	case oapi.GenderFemale:
		return 2
	case oapi.GenderOther:
		return 3
	default:
		// 理论上不该出现，记录一下方便排查脏数据
		logger.L.Warn("unexpected gender value", zap.String("gender", string(g)))
		return 0
	}
}

func UpdateUser(
	ctx context.Context,
	app *config.App,
	req oapi.UpdateUserRequestObject,
) (*oapi.UserPrivate, error) {
	userClaims, ok := middleware.GetUserFromContextOnly(ctx)
	if !ok {
		return nil, common.ErrUserNotInContext
	}

	// 获取目标用户信息
	userInfo, err := user_repo.GetUserInfoByID(ctx, app.DB, userClaims.UserID)
	if err != nil {
		return nil, err
	}

	// 构造用户表的部分更新参数
	userParams := dao.UpdateUserParams{
		Nickname:  &req.Body.Nickname,
		AvatarURL: &req.Body.AvatarUrl,
		Language:  &req.Body.Language,
		Biography: &req.Body.Bio,
	}

	// 构造 profile 的部分更新参数
	profileParams := dao.UpdateUserProfileParams{
		Country: req.Body.Country,
	}
	if req.Body.Gender != nil {
		genderVal := DTOToGender(*req.Body.Gender)
		profileParams.Gender = &genderVal
	}
	if req.Body.Birthday != nil {
		birthday := req.Body.Birthday.Time
		profileParams.Birthday = &birthday
	}

	privaciesParams := dao.UpdateUserPrivaciesParams{
		EmailVisibility:    nil,
		BirthdayVisibility: nil,
		GenderVisibility:   nil,
		CountryVisibility:  nil,
	}

	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.WrapPrefix(err, "failed to begin transaction", 0)
	}
	defer tx.Rollback()

	updatedUserInfo, err := user_repo.UpdateUser(ctx, tx, userInfo.ID, userParams)
	if err != nil {
		return nil, err
	}
	updatedProfile, err := user_repo.UpdateUserProfile(ctx, tx, userInfo.ID, profileParams)
	if err != nil {
		return nil, err
	}
	updatedPrivacies, err := user_repo.UpdateUserPrivacies(ctx, tx, userInfo.ID, privaciesParams)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.WrapPrefix(err, "failed to commit transaction", 0)
	}

	stats, err := user_repo.GetUserStatisticsByID(ctx, tx, userInfo.ID)
	if err != nil {
		return nil, err
	}
	logins, err := user_repo.GetLatestLoginLogByID(ctx, tx, userInfo.ID)
	if err != nil {
		return nil, err
	}

	res := transformer.UserModelToPrivate(
		*updatedUserInfo,
		*updatedProfile,
		*updatedPrivacies,
		*stats,
		*logins,
	)
	return &res, nil
}
