package services

import (
	"context"
	"genshin-quiz/config"
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/oapi"
	user_repo "genshin-quiz/internal/repository/user"
	"genshin-quiz/internal/transformer"
	"time"

	api_error "genshin-quiz/internal/errors"

	"github.com/go-errors/errors"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/golang-jwt/jwt/v5"
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
	if err != nil {
		return nil, err
	}
	if user != nil {
		return nil, api_error.ErrUserAlreadyExists
	}

	// 创建用户
	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.WrapPrefix(err, "failed to begin transaction", 0)
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
		return nil, errors.WrapPrefix(err, "failed to commit transaction", 0)
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
	ip, _ := ctx.Value("real_ip").(string)
	// 检测用户是否存在
	user, err := user_repo.GetUserByEmail(ctx, app.DB, string(email))
	if err != nil {
		return nil, err
	}

	// 验证密码
	err = user_repo.CheckPassword(ctx, app.DB, user.ID, pwd)
	if err != nil {
		return nil, err
	}

	// 登录流程
	return realLogin(ctx, app.DB, app.Config.JWTSecret, user, ip)
}

func realLogin(
	ctx context.Context,
	db qrm.DB,
	JWTSecret string,
	res *model.Users,
	ip string,
) (*oapi.AuthResponse, error) {
	// 生成 JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": res.ID,
		"email":   res.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(JWTSecret))
	if err != nil {
		return nil, err
	}

	// 写登录日志

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
