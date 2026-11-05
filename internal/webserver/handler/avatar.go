package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"genshin-quiz/internal/dao"
	user_repo "genshin-quiz/internal/repository/user"
	mw "genshin-quiz/internal/webserver/middleware"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const (
	maxAvatarSize   = 2 << 20
	avatarURLPrefix = "/auth/avatars/"
)

var (
	avatarExtensions = map[string]string{
		"image/png":  ".png",
		"image/jpeg": ".jpg",
		"image/gif":  ".gif",
		"image/webp": ".webp",
	}
	avatarFilePattern = regexp.MustCompile(`^[0-9]+-[0-9a-f]{16}\.(png|jpg|gif|webp)$`)
)

func avatarDir() string {
	if dir := os.Getenv("AVATAR_DIR"); dir != "" {
		return dir
	}
	return "./data/avatars"
}

// avatarPublicURL 返回可被浏览器直接访问的完整地址，基于 OAuth 回调所用的对外域名。
func avatarPublicURL(r *http.Request, name string) string {
	base := strings.TrimRight(os.Getenv("OAUTH_CALLBACK_BASE_URL"), "/")
	if base == "" {
		scheme := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		base = scheme + "://" + r.Host
	}
	return base + avatarURLPrefix + name
}

func avatarFileFromURL(url string) (string, bool) {
	idx := strings.LastIndex(url, avatarURLPrefix)
	if idx < 0 {
		return "", false
	}
	name := url[idx+len(avatarURLPrefix):]
	return name, avatarFilePattern.MatchString(name)
}

func (h *Handler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	claims, ok := mw.GetUserFromContext(r)
	if !ok {
		writeNodeAPIError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarSize+(64<<10))
	file, _, err := r.FormFile("avatar")
	if err != nil {
		writeNodeAPIError(w, http.StatusBadRequest, "请选择不超过 2MB 的图片文件")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxAvatarSize+1))
	if err != nil || len(data) == 0 || len(data) > maxAvatarSize {
		writeNodeAPIError(w, http.StatusBadRequest, "头像大小不能超过 2MB")
		return
	}
	ext, allowed := avatarExtensions[http.DetectContentType(data)]
	if !allowed {
		writeNodeAPIError(w, http.StatusBadRequest, "仅支持 PNG、JPG、GIF、WebP 图片")
		return
	}

	suffix := make([]byte, 8)
	if _, err := rand.Read(suffix); err != nil {
		writeNodeAPIError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	name := fmt.Sprintf("%d-%s%s", claims.UserID, hex.EncodeToString(suffix), ext)

	dir := avatarDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		h.app.Logger.Error("Failed to create avatar dir", zap.Error(err))
		writeNodeAPIError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
		h.app.Logger.Error("Failed to save avatar", zap.Error(err))
		writeNodeAPIError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	previous, err := user_repo.GetUserInfoByID(r.Context(), h.app.DB, claims.UserID)
	if err != nil {
		_ = os.Remove(filepath.Join(dir, name))
		h.app.Logger.Error("Failed to load user for avatar", zap.Error(err))
		writeNodeAPIError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	url := avatarPublicURL(r, name)
	if _, err := user_repo.UpdateUser(r.Context(), h.app.DB, claims.UserID, dao.UpdateUserParams{AvatarURL: &url}); err != nil {
		_ = os.Remove(filepath.Join(dir, name))
		h.app.Logger.Error("Failed to update avatar url", zap.Error(err))
		writeNodeAPIError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// 只清理本系统上传的旧头像，第三方登录带来的外链头像不动
	if previous.AvatarURL != nil {
		if old, ok := avatarFileFromURL(*previous.AvatarURL); ok && old != name {
			_ = os.Remove(filepath.Join(dir, old))
		}
	}

	writeNodeAPIJSON(w, http.StatusOK, map[string]string{"avatar_url": url})
}

func (h *Handler) ServeAvatar(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if !avatarFilePattern.MatchString(name) {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")
	http.ServeFile(w, r, filepath.Join(avatarDir(), name))
}
