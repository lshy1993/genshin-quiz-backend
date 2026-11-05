package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"genshin-quiz/config"
	"genshin-quiz/internal/common"
	user_repo "genshin-quiz/internal/repository/user"

	"github.com/go-errors/errors"
	"golang.org/x/crypto/bcrypt"
)

type LinkedIdentity struct {
	Provider string    `json:"provider"`
	Email    string    `json:"email,omitempty"`
	Name     string    `json:"name,omitempty"`
	LinkedAt time.Time `json:"linked_at"`
}

type IdentityList struct {
	Identities  []LinkedIdentity `json:"identities"`
	HasPassword bool             `json:"has_password"`
}

type identityExtra struct {
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

func isSocialProvider(provider string) bool {
	return provider == "google" || provider == "github"
}

func ListIdentities(ctx context.Context, app *config.App, userID int64) (*IdentityList, error) {
	rows, err := app.DB.QueryContext(ctx,
		`SELECT identity_type, extra_data, created_at FROM user_credentials WHERE user_id = $1 ORDER BY created_at`,
		userID,
	)
	if err != nil {
		return nil, errors.WrapPrefix(err, "list identities failed", 0)
	}
	defer rows.Close()

	result := &IdentityList{Identities: []LinkedIdentity{}}
	for rows.Next() {
		var identityType string
		var extra []byte
		var createdAt time.Time
		if err := rows.Scan(&identityType, &extra, &createdAt); err != nil {
			return nil, errors.WrapPrefix(err, "read identity failed", 0)
		}
		if identityType == "password" {
			result.HasPassword = true
			continue
		}
		if !isSocialProvider(identityType) {
			continue
		}
		item := LinkedIdentity{Provider: identityType, LinkedAt: createdAt}
		if len(extra) > 0 {
			var parsed identityExtra
			if json.Unmarshal(extra, &parsed) == nil {
				item.Email, item.Name = parsed.Email, parsed.Name
			}
		}
		result.Identities = append(result.Identities, item)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.WrapPrefix(err, "read identities failed", 0)
	}
	return result, nil
}

// LinkIdentity 把已验证的第三方身份绑定到当前用户；同一平台只允许绑定一个账号。
func LinkIdentity(ctx context.Context, app *config.App, userID int64, identity OAuthIdentity) error {
	if !isSocialProvider(string(identity.Provider)) || identity.Subject == "" {
		return common.NewBadRequestError("Unsupported identity provider")
	}

	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return errors.WrapPrefix(err, "begin link transaction failed", 0)
	}
	defer tx.Rollback()

	var owner int64
	err = tx.QueryRowContext(ctx,
		`SELECT user_id FROM user_credentials WHERE identity_type = $1 AND identifier = $2`,
		string(identity.Provider), identity.Subject,
	).Scan(&owner)
	switch {
	case err == nil && owner == userID:
		return nil
	case err == nil:
		return common.NewConflictError("该第三方账号已被其他用户绑定")
	case err != sql.ErrNoRows:
		return errors.WrapPrefix(err, "lookup identity failed", 0)
	}

	var exists bool
	if err := tx.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM user_credentials WHERE user_id = $1 AND identity_type = $2)`,
		userID, string(identity.Provider),
	).Scan(&exists); err != nil {
		return errors.WrapPrefix(err, "lookup provider binding failed", 0)
	}
	if exists {
		return common.NewConflictError("当前账号已绑定该平台的其他账号，请先解绑")
	}

	extra, err := json.Marshal(identityExtra{Email: identity.Email, Name: identity.Name})
	if err != nil {
		return errors.WrapPrefix(err, "encode identity failed", 0)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO user_credentials (user_id, identity_type, identifier, extra_data) VALUES ($1, $2, $3, $4)`,
		userID, string(identity.Provider), identity.Subject, extra,
	); err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return common.NewConflictError("该第三方账号已被其他用户绑定")
		}
		return errors.WrapPrefix(err, "link identity failed", 0)
	}
	if err := tx.Commit(); err != nil {
		return errors.WrapPrefix(err, "commit link transaction failed", 0)
	}
	return nil
}

// UnlinkIdentity 解绑第三方账号，但必须保证用户至少还剩一种登录方式。
func UnlinkIdentity(ctx context.Context, app *config.App, userID int64, provider string) error {
	if !isSocialProvider(provider) {
		return common.NewBadRequestError("Unsupported identity provider")
	}

	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return errors.WrapPrefix(err, "begin unlink transaction failed", 0)
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx,
		`SELECT identity_type FROM user_credentials WHERE user_id = $1 FOR UPDATE`, userID)
	if err != nil {
		return errors.WrapPrefix(err, "lock identities failed", 0)
	}
	total, found := 0, false
	for rows.Next() {
		var identityType string
		if err := rows.Scan(&identityType); err != nil {
			rows.Close()
			return errors.WrapPrefix(err, "read identity failed", 0)
		}
		total++
		if identityType == provider {
			found = true
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return errors.WrapPrefix(err, "read identities failed", 0)
	}

	if !found {
		return common.NewNotFoundError("该平台尚未绑定")
	}
	if total <= 1 {
		return common.NewBadRequestError("至少需要保留一种登录方式")
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM user_credentials WHERE user_id = $1 AND identity_type = $2`, userID, provider,
	); err != nil {
		return errors.WrapPrefix(err, "unlink identity failed", 0)
	}
	if err := tx.Commit(); err != nil {
		return errors.WrapPrefix(err, "commit unlink transaction failed", 0)
	}
	return nil
}

const (
	minPasswordLength = 6
	// bcrypt 只处理前 72 字节，超出会被静默截断
	maxPasswordBytes = 72
)

// SetPassword 为还没有密码的账号（例如仅通过第三方登录创建的账号）设置登录密码，
// 密码凭证的 identifier 与注册时一致，使用账号邮箱。
func SetPassword(ctx context.Context, app *config.App, userID int64, newPassword string) error {
	if len(newPassword) < minPasswordLength {
		return common.NewBadRequestError("密码长度不能少于 6 位")
	}
	if len(newPassword) > maxPasswordBytes {
		return common.NewBadRequestError("密码不能超过 72 字节")
	}

	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return errors.WrapPrefix(err, "begin set password transaction failed", 0)
	}
	defer tx.Rollback()

	user, err := user_repo.GetUserInfoByID(ctx, tx, userID)
	if err != nil {
		return err
	}

	var exists bool
	if err := tx.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM user_credentials WHERE user_id = $1 AND identity_type = 'password')`,
		userID,
	).Scan(&exists); err != nil {
		return errors.WrapPrefix(err, "lookup password credential failed", 0)
	}
	if exists {
		return common.NewConflictError("账号已设置密码，请使用“修改密码”")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.WrapPrefix(err, "hash password failed", 0)
	}
	hashedStr := string(hashed)
	if err := user_repo.InsertUserAuth(ctx, tx, userID, "password", user.Email, &hashedStr); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return errors.WrapPrefix(err, "commit set password transaction failed", 0)
	}
	return nil
}
