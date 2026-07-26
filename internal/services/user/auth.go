package services

import (
	"context"
	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/enum"
	user_repo "genshin-quiz/internal/repository/user"
	"genshin-quiz/internal/webserver/middleware"
	"log"
	"time"

	"genshin-quiz/internal/common"

	"genshin-quiz/internal/util"

	"github.com/go-errors/errors"

	"golang.org/x/crypto/bcrypt"
)

func ForgotPassword(
	ctx context.Context,
	app *config.App,
	req oapi.PostForgotPasswordRequestObject,
) (*oapi.PostForgotPassword200Response, error) {
	email := string(req.Body.Email)

	// 检测用户是否存在
	user, err := user_repo.GetUserByEmail(ctx, app.DB, email)
	// 必须存在
	if err != nil {
		return nil, err
	}

	// 生成临时的token
	rawToken, err := user_repo.InsertUserToken(
		ctx,
		app.DB,
		user.ID,
		enum.TokenTypePasswordReset,
		24*time.Hour,
	)
	if err != nil {
		return nil, err
	}

	// 发送给邮箱
	finalURL := util.GenerateResetLink(app.Config.Domain, rawToken)
	log.Println(finalURL)
	err = app.SendEmail(email, "Email verification", finalURL)
	if err != nil {
		return nil, err
	}

	return &oapi.PostForgotPassword200Response{}, nil
}

func ResetPassword(
	ctx context.Context,
	app *config.App,
	req oapi.PostResetPasswordRequestObject,
) (*oapi.PostResetPassword200Response, error) {
	rawToken := req.Body.Token
	newPassword := req.Body.Password

	// 1. 开启数据库事务
	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.WrapPrefix(err, "begin transaction failed", 0)
	}
	defer tx.Rollback() // 确保出现异常或失败时自动回滚

	// 2. ⚡️ 核销 Token（传入事务 tx，验证 Token 是否有效并标记为已使用 is_used = true）
	tokenRecord, err := user_repo.VerifyAndUseToken(
		ctx,
		tx,
		rawToken,
		enum.TokenTypePasswordReset,
	)
	if err != nil {
		// Token 无效、已过期或已被使用
		return nil, common.ErrInvalidToken
	}

	// 3. 将用户提交的新密码进行 bcrypt 哈希加密
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.WrapPrefix(err, "hash new password failed", 0)
	}

	// 4. 更新用户的密码凭证（根据 tokenRecord.UserID 更新 user_credentials 表）
	hashedPwdStr := string(hashedPwd)
	err = user_repo.UpdateUserCredential(
		ctx,
		tx,
		tokenRecord.UserID,
		"password",
		&hashedPwdStr,
	)
	if err != nil {
		return nil, errors.WrapPrefix(err, "update user password failed", 0)
	}

	// 5. 提交事务
	if err := tx.Commit(); err != nil {
		return nil, errors.WrapPrefix(err, "commit transaction failed", 0)
	}

	return &oapi.PostResetPassword200Response{}, nil
}

func ChangePassword(
	ctx context.Context,
	app *config.App,
	req oapi.PostChangePasswordRequestObject,
) (*oapi.PostChangePassword200Response, error) {
	// 从 context 中获取用户信息
	userClaims, ok := middleware.GetUserFromContextOnly(ctx)
	if !ok {
		return nil, common.ErrUserNotInContext
	}

	userID := userClaims.UserID
	oldPwd := req.Body.OldPassword
	newPwd := req.Body.NewPassword

	// 2. 检查新旧密码是否一致
	if oldPwd == newPwd {
		return nil, common.ErrInvalidCredentials
	}

	// 3. 获取用户当前的密码凭证 (identity_type = 'password')
	cred, err := user_repo.GetUserCredential(ctx, app.DB, userID, "password")
	if err != nil {
		return nil, err
	}

	// 4. 校验旧密码是否正确
	err = bcrypt.CompareHashAndPassword([]byte(*cred.Credential), []byte(oldPwd))
	if err != nil {
		return nil, common.ErrInvalidCredentials
	}

	// 5. 对新密码进行 bcrypt 哈希加密
	hashedNewPwd, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.WrapPrefix(err, "hash new password failed", 0)
	}

	// 6. 更新数据库中的密码凭证
	hashedNewPwdStr := string(hashedNewPwd)
	err = user_repo.UpdateUserCredential(ctx, app.DB, userID, "password", &hashedNewPwdStr)
	if err != nil {
		return nil, errors.WrapPrefix(err, "update password failed", 0)
	}

	return &oapi.PostChangePassword200Response{}, nil
}

func SendVerificationEmail(
	ctx context.Context,
	app *config.App,
	req oapi.PostSendVerificationEmailRequestObject,
) (*oapi.PostSendVerificationEmail200Response, error) {
	email := string(req.Body.Email)

	// 检测用户是否存在
	user, err := user_repo.GetUserByEmail(ctx, app.DB, email)
	// 必须存在
	if err != nil {
		return nil, err
	}

	// 生成临时的token
	rawToken, err := user_repo.InsertUserToken(
		ctx,
		app.DB,
		user.ID,
		enum.TokenTypeEmailVerify,
		24*time.Hour,
	)
	if err != nil {
		return nil, err
	}
	// 发送给邮箱
	finalURL := util.GenerateEmailVerifyLink(app.Config.Domain, rawToken)
	log.Println(finalURL)
	err = app.SendEmail(email, "Email verification", finalURL)
	if err != nil {
		return nil, err
	}

	return &oapi.PostSendVerificationEmail200Response{}, nil
}

func VerifyEmail(
	ctx context.Context,
	app *config.App,
	req oapi.PostVerifyEmailRequestObject,
) (*oapi.PostVerifyEmail200Response, error) {
	rawToken := req.Body.Token

	// 1. 开启数据库事务
	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.WrapPrefix(err, "begin transaction failed", 0)
	}
	defer tx.Rollback() // 确保出现异常或失败时自动回滚

	// 2. ⚡️ 核销 Token（传入事务 tx，验证 Token 是否有效并标记为已使用 is_used = true）
	tokenRecord, err := user_repo.VerifyAndUseToken(
		ctx,
		tx,
		rawToken,
		enum.TokenTypeEmailVerify,
	)
	if err != nil {
		// Token 无效、已过期或已被使用
		return nil, common.ErrInvalidToken
	}
	// 3.标记用户邮箱已验证
	log.Println(tokenRecord.TokenHash)

	return &oapi.PostVerifyEmail200Response{}, nil
}
