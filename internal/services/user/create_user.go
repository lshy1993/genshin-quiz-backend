package services

import (
	"context"
	"errors"
	"genshin-quiz/config"
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/dao/transformer"
	user_repo "genshin-quiz/internal/repository/user"
	"genshin-quiz/internal/webserver/middleware"

	"genshin-quiz/internal/common"

	go_errors "github.com/go-errors/errors"
	"github.com/go-jet/jet/v2/qrm"
	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(
	ctx context.Context,
	app *config.App,
	req oapi.PostRegisterUserRequestObject,
) (*oapi.AuthResponse, error) {
	email := req.Body.Email
	pwd := req.Body.Password

	// 检测用户是否存在
	user, err := user_repo.GetUserByEmail(ctx, app.DB, string(email))
	if err != nil && !errors.Is(err, common.ErrUserNotFound) {
		// 其他错误
		return nil, err
	} else if user != nil {
		return nil, common.ErrUserAlreadyExists
	}

	// 创建用户
	tx, err := app.DB.Begin()
	if err != nil {
		return nil, go_errors.WrapPrefix(err, "failed to begin transaction", 0)
	}
	res, err := user_repo.InsertUser(ctx, tx, string(email))
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = user_repo.InsertUserAuth(ctx, tx, res.ID, pwd)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	response, err := realLogin(ctx, tx, app.Config.JWTSecret, res)
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

func LoginUser(
	ctx context.Context,
	app *config.App,
	req oapi.PostLoginUserRequestObject,
) (*oapi.AuthResponse, error) {
	email := req.Body.Email
	pwd := req.Body.Password

	// 获取用户信息
	authInfo, err := user_repo.GetPasswordByEmail(ctx, app.DB, string(email))
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(authInfo.Auth.PasswordHash), []byte(pwd))
	if err != nil {
		// 密码错误
		return nil, common.ErrInvalidCredentials
	}

	// 登录流程
	return realLogin(ctx, app.DB, app.Config.JWTSecret, &authInfo.User)
}

func realLogin(
	ctx context.Context,
	db qrm.DB,
	secret string,
	res *model.Users,
) (*oapi.AuthResponse, error) {
	// 生成 JWT
	tokenString, err := middleware.GenerateJWT(res.ID, res.Email, secret)
	if err != nil {
		return nil, err
	}

	// 写登录日志
	ip, _ := ctx.Value("real_ip").(string)
	loginInfo, err := user_repo.InsertLoginLog(ctx, db, res.ID, ip)
	if err != nil {
		return nil, err
	}
	// TODO:获取用户的其他统计信息

	return &oapi.AuthResponse{
		Token: tokenString,
		User:  transformer.UserModelToDTO(*res, *loginInfo),
	}, nil
}
