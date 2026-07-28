package services

import (
	"context"
	"genshin-quiz/config"
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/dao/transformer"
	"genshin-quiz/internal/enum"
	user_repo "genshin-quiz/internal/repository/user"
	"genshin-quiz/internal/webserver/middleware"

	"genshin-quiz/internal/common"

	"github.com/go-errors/errors"
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
		return nil, errors.WrapPrefix(err, "login user failed", 0)
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

	// TODO:获取用户的其他统计信息
	userID := authInfo.Users.ID
	userProfile, err := user_repo.GetUserProfileByID(ctx, app.DB, userID)
	if err != nil {
		return nil, errors.WrapPrefix(err, "login user failed", 0)
	}
	userPrivacies, err := user_repo.GetUserPrivaciesByID(ctx, app.DB, userID)
	if err != nil {
		return nil, errors.WrapPrefix(err, "login user failed", 0)
	}
	userStats, err := user_repo.GetUserStatisticsByID(ctx, app.DB, userID)
	if err != nil {
		return nil, errors.WrapPrefix(err, "login user failed", 0)
	}

	// 登录流程
	return realLogin(
		ctx,
		app.DB,
		app.Config.JWTSecret,
		"password",
		&authInfo.Users,
		userProfile,
		userPrivacies,
		userStats,
	)
}

func realLogin(
	ctx context.Context,
	db qrm.DB,
	secret string,
	loginType enum.LoginProvider, // "password", "google"
	user *model.Users,
	profile *model.UserProfiles,
	privacies *model.UserPrivacies,
	stats *model.UserStats,
) (*oapi.AuthResponse, error) {
	// 生成 JWT
	tokenString, err := middleware.GenerateJWT(user.ID, user.Email, secret)
	if err != nil {
		return nil, errors.WrapPrefix(err, "GenerateJWT failed", 0)
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

	// 写登录日志
	loginInfo, err := user_repo.InsertLoginLog(
		ctx,
		db,
		user.ID,
		ip,
		userAgent,
		loginType,
		enum.LoginStatusSuccess,
	)
	if err != nil {
		return nil, err
	}

	return &oapi.AuthResponse{
		Token: tokenString,
		User: transformer.UserModelToPrivate(
			*user,
			*profile,
			*privacies,
			*stats,
			*loginInfo,
		),
	}, nil
}
