// Package bootstrap 负责首次部署时确定初始管理员。
//
// 使用邮箱而不是 user_id：user_id 在换库、重建、迁移后都会变化，邮箱是稳定的业务标识。
// 规则：
//   - 只有在系统里一个管理员都没有时，启动引导才会生效
//   - 邮箱已存在：只提权，不修改密码
//   - 邮箱不存在：必须提供初始密码，创建管理员账号
//   - 显式 CLI 可以跳过“没有管理员”限制，用于给已有系统增加管理员
package bootstrap

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"genshin-quiz/config"
	"genshin-quiz/internal/common"
	user_repo "genshin-quiz/internal/repository/user"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const minAdminPasswordLength = 12

// AdminCount 返回当前 user_role=1 的用户数量。
func AdminCount(ctx context.Context, db *sql.DB) (int, error) {
	var count int
	if err := db.QueryRowContext(ctx, `SELECT count(*)::int FROM users WHERE user_role = 1`).Scan(&count); err != nil {
		return 0, fmt.Errorf("count admins: %w", err)
	}
	return count, nil
}

// EnsureInitialAdmin 只在系统没有任何管理员时创建/提权初始管理员。
func EnsureInitialAdmin(
	ctx context.Context,
	app *config.App,
	email, password, nickname string,
	resetPassword bool,
) error {
	count, err := AdminCount(ctx, app.DB)
	if err != nil {
		return err
	}
	if count > 0 {
		app.Logger.Info("Bootstrap admin skipped: an admin already exists", zap.Int("admin_count", count))
		return nil
	}
	return ensureAdmin(ctx, app, email, password, nickname, resetPassword)
}

// EnsureAdmin 显式提权/创建管理员；即使已有其他管理员也会执行。
func EnsureAdmin(
	ctx context.Context,
	app *config.App,
	email, password, nickname string,
	resetPassword bool,
) error {
	return ensureAdmin(ctx, app, email, password, nickname, resetPassword)
}

func ensureAdmin(
	ctx context.Context,
	app *config.App,
	email, password, nickname string,
	resetPassword bool,
) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || !strings.Contains(email, "@") {
		return fmt.Errorf("invalid admin email: %q", email)
	}

	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin admin bootstrap transaction: %w", err)
	}
	defer tx.Rollback()

	user, err := user_repo.GetUserByEmail(ctx, tx, email)
	if err != nil && !errors.Is(err, common.ErrUserNotFound) {
		return fmt.Errorf("find admin by email: %w", err)
	}

	var userID int64
	if user == nil {
		if password == "" {
			return fmt.Errorf("admin %s does not exist; provide BOOTSTRAP_ADMIN_PASSWORD or register the account first", email)
		}
		if len(password) < minAdminPasswordLength {
			return fmt.Errorf("bootstrap admin password must be at least %d characters", minAdminPasswordLength)
		}
		if len(password) > 72 {
			return fmt.Errorf("bootstrap admin password must be at most 72 bytes")
		}

		hashed, hashErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if hashErr != nil {
			return fmt.Errorf("hash bootstrap admin password: %w", hashErr)
		}
		hashedStr := string(hashed)

		created, createErr := user_repo.InsertUser(ctx, tx, email, "zh")
		if createErr != nil {
			return fmt.Errorf("create bootstrap admin: %w", createErr)
		}
		userID = created.ID

		if authErr := user_repo.InsertUserAuth(ctx, tx, userID, "password", email, &hashedStr); authErr != nil {
			return fmt.Errorf("create bootstrap admin credential: %w", authErr)
		}
		if _, profileErr := user_repo.InsertUserProfile(ctx, tx, userID); profileErr != nil {
			return fmt.Errorf("create bootstrap admin profile: %w", profileErr)
		}
		if _, privacyErr := user_repo.InsertUserPrivacies(ctx, tx, userID); privacyErr != nil {
			return fmt.Errorf("create bootstrap admin privacy: %w", privacyErr)
		}
		if _, statsErr := user_repo.InsertUserStats(ctx, tx, userID); statsErr != nil {
			return fmt.Errorf("create bootstrap admin stats: %w", statsErr)
		}
	} else {
		if user.Status != 0 || user.DeletedAt != nil {
			return fmt.Errorf("admin account %s is not active", email)
		}
		userID = user.ID

		if resetPassword {
			if password == "" {
				return fmt.Errorf("reset password requested but no password was provided")
			}
			if err := upsertPassword(ctx, tx, userID, email, password); err != nil {
				return err
			}
		} else if password != "" {
			current, err := user_repo.GetPasswordByEmail(ctx, tx, email)
			if err != nil {
				return fmt.Errorf("load existing admin credential: %w", err)
			}
			if current.Credential != nil {
				if bcrypt.CompareHashAndPassword([]byte(*current.Credential), []byte(password)) != nil {
					return fmt.Errorf(
						"email %s already exists and the password does not match; use -reset-password to take it over",
						email,
					)
				}
			} else if !user.EmailVerified {
				return fmt.Errorf("email %s is not verified; provide -reset-password to take it over", email)
			}
		} else if !user.EmailVerified {
			return fmt.Errorf("email %s is not verified; provide a password or use -reset-password", email)
		}
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE users
		 SET user_role = 1,
		     nickname = COALESCE(NULLIF($2, ''), nickname),
		     email_verified = TRUE,
		     updated_at = NOW()
		 WHERE id = $1`,
		userID, strings.TrimSpace(nickname),
	)
	if err != nil {
		return fmt.Errorf("promote bootstrap admin: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit bootstrap admin: %w", err)
	}

	app.Logger.Info("Bootstrap admin ready",
		zap.Int64("user_id", userID),
		zap.String("email", email),
	)
	return nil
}

func upsertPassword(ctx context.Context, tx *sql.Tx, userID int64, email, password string) error {
	if len(password) < minAdminPasswordLength {
		return fmt.Errorf("bootstrap admin password must be at least %d characters", minAdminPasswordLength)
	}
	if len(password) > 72 {
		return fmt.Errorf("bootstrap admin password must be at most 72 bytes")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash bootstrap admin password: %w", err)
	}
	_, err = tx.ExecContext(ctx,
		`INSERT INTO user_credentials (user_id, identity_type, identifier, credential, created_at, updated_at)
		 VALUES ($1, 'password', $2, $3, NOW(), NOW())
		 ON CONFLICT (identity_type, identifier)
		 DO UPDATE SET user_id = $1, credential = $3, updated_at = NOW()`,
		userID, email, string(hashed),
	)
	if err != nil {
		return fmt.Errorf("upsert bootstrap admin password: %w", err)
	}
	return nil
}
