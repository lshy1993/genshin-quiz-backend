package services

import (
	"context"
	"errors"
	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	user_repo "genshin-quiz/internal/repository/user"
	"genshin-quiz/internal/util"

	"genshin-quiz/internal/common"

	go_errors "github.com/go-errors/errors"
	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(
	ctx context.Context,
	app *config.App,
	req oapi.PostRegisterUserRequestObject,
) (*oapi.AuthResponse, error) {
	email := req.Body.Email
	pwd := req.Body.Password
	language := req.Body.Language

	// 检测用户是否存在
	user, err := user_repo.GetUserByEmail(ctx, app.DB, string(email))
	if err != nil && !errors.Is(err, common.ErrUserNotFound) {
		// 其他错误
		return nil, err
	} else if user != nil {
		return nil, common.ErrUserAlreadyExists
	}

	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, go_errors.WrapPrefix(err, "failed to begin transaction", 0)
	}
	defer tx.Rollback()

	// 创建用户
	res, err := user_repo.InsertUser(ctx, tx, string(email), util.LanguageOrDefault(language))
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return nil, go_errors.WrapPrefix(err, "hash password failed", 0)
	}
	hashedPwdStr := string(hashedPwd)
	userID := res.ID
	// use new auth
	err = user_repo.InsertUserAuth(ctx, tx, userID, "password", string(email), &hashedPwdStr)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	// 创建关联表的行
	profile, err := user_repo.InsertUserProfile(ctx, tx, userID)
	if err != nil {
		tx.Rollback()
		return nil, go_errors.WrapPrefix(err, "failed to insert user profile", 0)
	}
	privacies, err := user_repo.InsertUserPrivacies(ctx, tx, userID)
	if err != nil {
		tx.Rollback()
		return nil, go_errors.WrapPrefix(err, "failed to insert user privacies", 0)
	}
	stats, err := user_repo.InsertUserStats(ctx, tx, userID)
	if err != nil {
		tx.Rollback()
		return nil, go_errors.WrapPrefix(err, "failed to insert user stats", 0)
	}

	response, err := realLogin(
		ctx,
		tx,
		app.Config.JWTSecret,
		"password",
		res,
		profile,
		privacies,
		stats,
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, go_errors.WrapPrefix(err, "failed to commit transaction", 0)
	}

	return response, nil
}
