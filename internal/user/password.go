package user

import (
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用 bcrypt 生成密码哈希，返回可存储的字符串
func HashPassword(plain string) (string, error) {
	if plain == "" {
		return "", errors.New("password empty")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CreatePasswordForUser 在 user_passwords 表中插入或更新密码条目
func CreatePasswordForUser(db *sql.DB, userID int64, plainPassword string) error {
	hash, err := HashPassword(plainPassword)
	if err != nil {
		return err
	}

	// 使用 UPSERT（Postgres 的 ON CONFLICT）来插入或更新
	_, err = db.Exec(`
		INSERT INTO user_passwords (user_id, password_hash, password_algorithm, created_at, updated_at)
		VALUES ($1, $2, 'bcrypt', now(), now())
		ON CONFLICT (user_id) DO UPDATE SET password_hash = EXCLUDED.password_hash, updated_at = now()
	`, userID, hash)
	return err
}

// GetPasswordHashByUser 获取指定用户的密码哈希
func GetPasswordHashByUser(db *sql.DB, userID int64) (string, error) {
	var hash string
	err := db.QueryRow(`SELECT password_hash FROM user_passwords WHERE user_id = $1`, userID).Scan(&hash)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return hash, nil
}

// VerifyPassword 验证用户提供的明文密码是否匹配存储的哈希
func VerifyPassword(db *sql.DB, userID int64, plainPassword string) (bool, error) {
	hash, err := GetPasswordHashByUser(db, userID)
	if err != nil {
		return false, err
	}
	if hash == "" {
		return false, errors.New("password not set")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plainPassword)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// UpdateLastLogin 更新用户的 last_login_at 字段（示例）
func UpdateLastLogin(db *sql.DB, userID int64) error {
	_, err := db.Exec(`UPDATE users SET last_login_at = $1 WHERE id = $2`, time.Now().UTC(), userID)
	return err
}
