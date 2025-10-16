package services

import (
	"context"
	"genshin-quiz/config"
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/oapi"
	user_repo "genshin-quiz/internal/repository/user"
	"genshin-quiz/internal/transformer"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func RegisterUser(
	ctx context.Context,
	app *config.App,
	req oapi.PostRegisterUserRequestObject,
) (*oapi.AuthResponse, error) {
	email := req.Body.Email
	pwd := req.Body.Password

	// 先创建profile
	res, err := user_repo.Insert(ctx, app.DB, string(email))
	if err != nil {
		return nil, err
	}
	// 创建用户密码
	err = user_repo.InsertAuth(ctx, app.DB, res.ID, pwd)
	if err != nil {
		return nil, err
	}

	return realLogin(ctx, app, res)
}

func LoginUser(
	ctx context.Context,
	app *config.App,
	req oapi.PostLoginUserRequestObject,
) (*oapi.AuthResponse, error) {
	email := req.Body.Email
	pwd := req.Body.Password
	// 验证密码
	res, err := user_repo.GetUserByEmailAndPassword(ctx, app.DB, string(email), pwd)
	if err != nil {
		return nil, err
	}
	// 登录流程
	return realLogin(ctx, app, res)
}

func realLogin(
	ctx context.Context,
	app *config.App,
	res *model.Users,
) (*oapi.AuthResponse, error) {
	// 生成 JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": res.ID,
		"email":   res.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(app.Config.JWTSecret))
	if err != nil {
		return nil, err
	}

	// 写登录日志
	ip := ""
	loginInfo, err := user_repo.InsertLoginLog(ctx, app.DB, res.ID, ip)
	if err != nil {
		return nil, err
	}
	// TODO:获取用户的其他统计信息

	return &oapi.AuthResponse{
		Token: tokenString,
		User:  transformer.UserModelToDTO(*res, *loginInfo),
	}, nil
}
