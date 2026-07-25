package services

import (
	"context"
	"errors"
	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	user_repo "genshin-quiz/internal/repository/user"

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

	// 创建用户
	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, go_errors.WrapPrefix(err, "failed to begin transaction", 0)
	}
	res, err := user_repo.InsertUser(ctx, tx, string(email), language)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return nil, go_errors.WrapPrefix(err, "hash password failed", 0)
	}
	hashedPwdStr := string(hashedPwd)
	// use new auth
	err = user_repo.InsertUserAuth(ctx, tx, res.ID, "password", string(email), &hashedPwdStr)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	response, err := realLogin(ctx, tx, app.Config.JWTSecret, res, "password")
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, go_errors.WrapPrefix(err, "failed to commit transaction", 0)
	}

	return response, nil
}
