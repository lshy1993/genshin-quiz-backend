package user_repo

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/db/genshinquiz/public/table"
	"genshin-quiz/internal/common"
	"genshin-quiz/internal/dao"
	"genshin-quiz/internal/enum"
	"time"

	"github.com/go-errors/errors"
	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
)

func loginProviderToInt16(provider enum.LoginProvider) (int16, error) {
	switch provider {
	case "password":
		return 0, nil
	case "google":
		return 1, nil
	case "apple":
		return 2, nil
	case "github":
		return 3, nil
	default:
		return 0, fmt.Errorf("%w: %s", common.ErrInvalidLoginProvider, provider)
	}
}

func InsertUserAuth(
	ctx context.Context,
	db qrm.DB,
	userID int64,
	identityType string, // 'password', 'google', 'github'
	identifier string, // email, openid, phone
	credential *string, // 密码哈希 (OAuth 可传 nil)
) error {
	tbl := table.UserCredentials
	now := time.Now()

	insertStmt := tbl.INSERT(
		tbl.UserID,
		tbl.IdentityType,
		tbl.Identifier,
		tbl.Credential,
		tbl.CreatedAt,
		tbl.UpdatedAt,
	).MODEL(model.UserCredentials{
		UserID:       userID,
		IdentityType: identityType,
		Identifier:   identifier,
		Credential:   credential,
		CreatedAt:    now,
		UpdatedAt:    now,
	})

	_, err := insertStmt.ExecContext(ctx, db)
	if err != nil {
		errStr := err.Error()
		if errStr != "" &&
			(contains(errStr, "duplicate key") || contains(errStr, "unique constraint")) {
			return common.NewBadRequestError("this identity is already linked to an account")
		}
		return errors.WrapPrefix(err, "insert user identity failed", 0)
	}

	return nil
}

func InsertLoginLog(
	ctx context.Context,
	db qrm.DB,
	userID int64,
	ip string,
	userAgent *string,
	loginType enum.LoginProvider, // "password", "google"
	status enum.LoginStatus, // "SUCCESS", "FAILED"
) (*model.UserLoginLogs, error) {
	tbl := table.UserLoginLogs

	// 1. 设置默认值
	credType, err := loginProviderToInt16(loginType)
	if err != nil {
		// 返回 400 Bad Request，而不是让非法数据流入数据库
		return nil, errors.WrapPrefix(err, "cred provider wrong", 0)
	}

	now := time.Now()
	insertStmt := tbl.INSERT(
		tbl.UserID,
		tbl.IPAddress,
		tbl.UserAgent,
		tbl.CredentialType,
		tbl.Status,
		tbl.LoginAt,
	).
		MODEL(model.UserLoginLogs{
			UserID:         userID,
			IPAddress:      ip,
			UserAgent:      userAgent,
			CredentialType: credType,
			Status:         int16(status),
			LoginAt:        now,
		}).
		RETURNING(tbl.AllColumns)

	var result model.UserLoginLogs
	err = insertStmt.QueryContext(ctx, db, &result)
	if err != nil {
		return nil, errors.WrapPrefix(err, "insert login logs error", 0)
	}
	return &result, nil
}

func GetLatestLoginLogByID(
	ctx context.Context,
	db qrm.DB,
	userID int64,
) (*model.UserLoginLogs, error) {
	tbl := table.UserLoginLogs

	stmt := pg.SELECT(tbl.AllColumns).
		FROM(tbl).
		WHERE(tbl.UserID.EQ(pg.Int64(userID))).
		ORDER_BY(tbl.LoginAt.DESC()).
		LIMIT(1)

	var result model.UserLoginLogs
	err := stmt.QueryContext(ctx, db, &result)
	if err != nil {
		return nil, errors.WrapPrefix(err, "get latest login log error", 0)
	}

	return &result, nil
}

func GetUserCredential(
	ctx context.Context,
	db qrm.DB,
	userID int64,
	identityType string,
) (*model.UserCredentials, error) {
	tbl := table.UserCredentials

	stmt := pg.SELECT(tbl.AllColumns).
		FROM(tbl).
		WHERE(
			tbl.UserID.EQ(pg.Int(userID)).
				AND(tbl.IdentityType.EQ(pg.String(identityType))),
		)

	var cred model.UserCredentials
	err := stmt.QueryContext(ctx, db, &cred)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, common.ErrUserNotFound
		}
		return nil, err
	}

	return &cred, nil
}

func GetPasswordByEmail(
	ctx context.Context,
	db qrm.DB,
	email string,
) (*dao.UserInfoWithAuth, error) {
	tbl := table.Users
	authTbl := table.UserCredentials
	stmt := pg.SELECT(
		tbl.AllColumns,
		authTbl.IdentityType,
		authTbl.Identifier,
		authTbl.Credential,
	).
		FROM(tbl.LEFT_JOIN(
			authTbl,
			tbl.ID.EQ(authTbl.UserID).
				AND(authTbl.IdentityType.EQ(pg.String("password"))), // 限定凭证类型为 password
		)).
		WHERE(
			tbl.Email.EQ(pg.String(email)),
		)
	var auth []dao.UserInfoWithAuth
	err := stmt.QueryContext(ctx, db, &auth)
	if err != nil {
		return nil, errors.WrapPrefix(err, "checking password failed", 0)
	}
	if len(auth) == 0 {
		return nil, common.ErrUserNotFound
	}

	return &auth[0], nil
}

func UpdateUserCredential(
	ctx context.Context,
	db qrm.DB,
	userID int64,
	identityType string,
	credential *string,
) error {
	tbl := table.UserCredentials

	stmt := tbl.UPDATE(tbl.Credential, tbl.UpdatedAt).
		SET(
			credential,
			time.Now(),
		).
		WHERE(
			tbl.UserID.EQ(pg.Int(userID)).
				AND(tbl.IdentityType.EQ(pg.String(identityType))),
		)

	_, err := stmt.ExecContext(ctx, db)
	if err != nil {
		return err
	}

	return nil
}

func InsertUserToken(
	ctx context.Context,
	db qrm.DB,
	userID int64,
	tokenType enum.TokenType, // "password_reset", "email_verification"
	expiresIn time.Duration,
) (string, error) {
	// 1. 生成 32 字节（256 bit）的密码学安全随机字符串作为原始 Token
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", errors.WrapPrefix(err, "generate random token failed", 0)
	}
	rawToken := hex.EncodeToString(randomBytes)

	// 2. 计算明文 Token 的 SHA-256 哈希用于数据库存储
	hashBytes := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hashBytes[:])

	// 3. 构建插入数据
	now := time.Now()
	expiresAt := now.Add(expiresIn)

	tbl := table.UserTokens

	insertStmt := tbl.INSERT(
		tbl.UserID,
		tbl.TokenType,
		tbl.TokenHash,
		tbl.IsUsed,
		tbl.ExpiresAt,
		tbl.CreatedAt,
	).
		MODEL(model.UserTokens{
			UserID:    userID,
			TokenType: tokenType.String(),
			TokenHash: tokenHash,
			IsUsed:    false,
			ExpiresAt: expiresAt,
			CreatedAt: now,
		})

	// 4. 执行插入
	_, err := insertStmt.ExecContext(ctx, db)
	if err != nil {
		return "", errors.WrapPrefix(err, "insert user token failed", 0)
	}

	// 5. 返回未加密的明文 rawToken（以便通过邮件/链接发给用户）
	return rawToken, nil
}

func VerifyAndUseToken(
	ctx context.Context,
	db qrm.DB,
	rawToken string,
	tokenType enum.TokenType, // "password_reset", "email_verification"
) (*model.UserTokens, error) {
	tbl := table.UserTokens

	// 1. 对传入的明文 rawToken 计算 SHA-256
	hashBytes := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hashBytes[:])

	// 2. 查询并更新中未过期且未使用的 Token
	now := time.Now()
	stmt := tbl.UPDATE().
		SET(
			tbl.IsUsed.SET(pg.Bool(true)),
		).
		WHERE(
			tbl.TokenHash.EQ(pg.String(tokenHash)).
				AND(tbl.TokenType.EQ(pg.String(tokenType.String()))).
				AND(tbl.IsUsed.IS_FALSE()).
				AND(tbl.ExpiresAt.GT(pg.TimestampzT(now))),
		).
		RETURNING(tbl.AllColumns)

	var tokenRecord model.UserTokens
	err := stmt.QueryContext(ctx, db, &tokenRecord)
	if err != nil {
		// 找不到token或过期
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, common.ErrInvalidToken
		}
		return nil, err
	}

	return &tokenRecord, nil
}
