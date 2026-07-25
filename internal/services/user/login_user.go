package services

import (
	"context"
	"genshin-quiz/config"
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/dao/transformer"
	user_repo "genshin-quiz/internal/repository/user"
	"genshin-quiz/internal/webserver/middleware"

	"genshin-quiz/internal/common"

	"github.com/go-jet/jet/v2/qrm"
	"golang.org/x/crypto/bcrypt"
)

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

	hashedPwd := authInfo.Credential
	if hashedPwd == nil {
		return nil, common.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(*hashedPwd), []byte(pwd))
	if err != nil {
		// 密码错误
		return nil, common.ErrInvalidCredentials
	}

	// 登录流程
	return realLogin(ctx, app.DB, app.Config.JWTSecret, &authInfo.Users, "password")
}

func realLogin(
	ctx context.Context,
	db qrm.DB,
	secret string,
	res *model.Users,
	loginType string, // "password", "google"
) (*oapi.AuthResponse, error) {
	// 生成 JWT
	tokenString, err := middleware.GenerateJWT(res.ID, res.Email, secret)
	if err != nil {
		return nil, err
	}
	// 从 Context 提取 IP 和 User-Agent
	ip, _ := ctx.Value(middleware.RealIPKey).(string)
	// 如果 IP 为空，赋值默认值，防止数据库 INET 字段解析报错
	if ip == "" {
		ip = "127.0.0.1"
	}

	var userAgent *string
	if ua, ok := ctx.Value(middleware.UserAgentKey).(string); ok && ua != "" {
		userAgent = &ua
	}

	if loginType == "" {
		loginType = "password"
	}

	// 写登录日志
	loginInfo, err := user_repo.InsertLoginLog(
		ctx,
		db,
		res.ID,
		ip,
		userAgent,
		&loginType,
		"SUCCESS",
	)
	if err != nil {
		return nil, err
	}
	// TODO:获取用户的其他统计信息

	return &oapi.AuthResponse{
		Token: tokenString,
		User:  transformer.UserModelToDTO(*res, *loginInfo),
	}, nil
}
