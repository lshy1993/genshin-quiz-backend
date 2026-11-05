package services

import (
	"context"
	"database/sql"
	"fmt"
	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/dao"
	"genshin-quiz/internal/enum"
	user_repo "genshin-quiz/internal/repository/user"
	"net/mail"
	"strings"

	"github.com/go-errors/errors"
)

type OAuthIdentity struct {
	Provider      enum.LoginProvider
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
	AvatarURL     string
}

func LoginWithOAuth(ctx context.Context, app *config.App, identity OAuthIdentity) (*oapi.AuthResponse, error) {
	identity.Email = strings.ToLower(strings.TrimSpace(identity.Email))
	identity.Subject = strings.TrimSpace(identity.Subject)
	if (identity.Provider != "google" && identity.Provider != "github") ||
		identity.Subject == "" || len(identity.Subject) > 255 || identity.Email == "" ||
		len(identity.Email) > 255 || !identity.EmailVerified {
		return nil, fmt.Errorf("provider did not return a verified email")
	}
	parsedEmail, err := mail.ParseAddress(identity.Email)
	if err != nil || parsedEmail.Address != identity.Email {
		return nil, fmt.Errorf("provider returned an invalid email")
	}

	var userID int64
	err = app.DB.QueryRowContext(ctx,
		`SELECT user_id FROM user_credentials WHERE identity_type = $1 AND identifier = $2`,
		string(identity.Provider), identity.Subject,
	).Scan(&userID)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.WrapPrefix(err, "lookup OAuth identity failed", 0)
	}

	if err == sql.ErrNoRows {
		tx, txErr := app.DB.BeginTx(ctx, nil)
		if txErr != nil {
			return nil, errors.WrapPrefix(txErr, "begin OAuth login transaction failed", 0)
		}
		defer tx.Rollback()

		accountRows, accountErr := tx.QueryContext(ctx,
			`SELECT id, status, deleted_at IS NOT NULL FROM users WHERE lower(email) = $1 ORDER BY id LIMIT 2`,
			identity.Email,
		)
		if accountErr != nil {
			return nil, errors.WrapPrefix(accountErr, "lookup OAuth email failed", 0)
		}
		var accountFound bool
		var status int16
		var deleted bool
		for accountRows.Next() {
			if accountFound {
				accountRows.Close()
				return nil, fmt.Errorf("multiple accounts match the verified email")
			}
			accountFound = true
			if scanErr := accountRows.Scan(&userID, &status, &deleted); scanErr != nil {
				accountRows.Close()
				return nil, errors.WrapPrefix(scanErr, "read OAuth email account failed", 0)
			}
		}
		if rowsErr := accountRows.Err(); rowsErr != nil {
			accountRows.Close()
			return nil, errors.WrapPrefix(rowsErr, "read OAuth email account failed", 0)
		}
		accountRows.Close()

		if !accountFound {
			user, createErr := user_repo.InsertUser(ctx, tx, identity.Email, "zh-CN")
			if createErr != nil {
				return nil, errors.WrapPrefix(createErr, "create OAuth user failed", 0)
			}
			userID = user.ID
			if _, createErr = user_repo.InsertUserProfile(ctx, tx, userID); createErr != nil {
				return nil, errors.WrapPrefix(createErr, "create OAuth user profile failed", 0)
			}
			if _, createErr = user_repo.InsertUserPrivacies(ctx, tx, userID); createErr != nil {
				return nil, errors.WrapPrefix(createErr, "create OAuth user privacy failed", 0)
			}
			if _, createErr = user_repo.InsertUserStats(ctx, tx, userID); createErr != nil {
				return nil, errors.WrapPrefix(createErr, "create OAuth user stats failed", 0)
			}

			name := strings.TrimSpace(identity.Name)
			if name == "" {
				name = strings.Split(identity.Email, "@")[0]
			}
			if len([]rune(name)) > 100 {
				name = string([]rune(name)[:100])
			}
			params := dao.UpdateUserParams{Nickname: &name}
			if strings.HasPrefix(identity.AvatarURL, "https://") {
				params.AvatarURL = &identity.AvatarURL
			}
			if _, createErr = user_repo.UpdateUser(ctx, tx, userID, params); createErr != nil {
				return nil, errors.WrapPrefix(createErr, "set OAuth user profile failed", 0)
			}
		} else if status != 0 || deleted {
			return nil, fmt.Errorf("account is unavailable")
		}

		if err = user_repo.InsertUserAuth(ctx, tx, userID, string(identity.Provider), identity.Subject, nil); err != nil {
			return nil, errors.WrapPrefix(err, "link OAuth identity failed", 0)
		}
		if err = user_repo.SetUserEmailVerified(ctx, tx, userID); err != nil {
			return nil, errors.WrapPrefix(err, "verify OAuth email failed", 0)
		}

		user, loadErr := user_repo.GetUserInfoByID(ctx, tx, userID)
		if loadErr != nil {
			return nil, loadErr
		}
		profile, loadErr := user_repo.GetUserProfileByID(ctx, tx, userID)
		if loadErr != nil {
			return nil, loadErr
		}
		privacies, loadErr := user_repo.GetUserPrivaciesByID(ctx, tx, userID)
		if loadErr != nil {
			return nil, loadErr
		}
		stats, loadErr := user_repo.GetUserStatisticsByID(ctx, tx, userID)
		if loadErr != nil {
			return nil, loadErr
		}
		response, loginErr := realLogin(ctx, tx, identity.Provider, user, profile, privacies, stats)
		if loginErr != nil {
			return nil, loginErr
		}
		if err = tx.Commit(); err != nil {
			return nil, errors.WrapPrefix(err, "commit OAuth login transaction failed", 0)
		}
		return response, nil
	}

	user, err := user_repo.GetUserInfoByID(ctx, app.DB, userID)
	if err != nil {
		return nil, err
	}
	if user.Status != 0 || user.DeletedAt != nil {
		return nil, fmt.Errorf("account is unavailable")
	}
	// 用户还没有头像时，用第三方头像补上；已有头像（含自行上传）不覆盖
	if (user.AvatarURL == nil || *user.AvatarURL == "") && strings.HasPrefix(identity.AvatarURL, "https://") {
		if updated, updateErr := user_repo.UpdateUser(ctx, app.DB, userID, dao.UpdateUserParams{AvatarURL: &identity.AvatarURL}); updateErr == nil {
			user = updated
		}
	}
	profile, err := user_repo.GetUserProfileByID(ctx, app.DB, userID)
	if err != nil {
		return nil, err
	}
	privacies, err := user_repo.GetUserPrivaciesByID(ctx, app.DB, userID)
	if err != nil {
		return nil, err
	}
	stats, err := user_repo.GetUserStatisticsByID(ctx, app.DB, userID)
	if err != nil {
		return nil, err
	}
	return realLogin(ctx, app.DB, identity.Provider, user, profile, privacies, stats)
}
