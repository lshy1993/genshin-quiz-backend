package handler

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"genshin-quiz/internal/enum"
	user_services "genshin-quiz/internal/services/user"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const oauthStateCookiePrefix = "oauth_state_"

const (
	oauthModeLink       = "link"
	oauthLinkPurpose    = "oauth_link"
	oauthLinkTokenTTL   = 5 * time.Minute
	maxIdentityBodySize = 1 << 14
)

type oauthState struct {
	Nonce     string `json:"nonce"`
	ReturnURL string `json:"return_url"`
	ExpiresAt int64  `json:"expires_at"`
	Mode      string `json:"mode,omitempty"`
}

// oauthLinkToken 是绑定模式回调后签发的短期凭证，只证明"某个第三方身份已通过验证"，
// 真正绑定时必须由已登录用户携带自己的 Bearer 令牌提交，因此无法被他人诱导绑定到其他账号。
type oauthLinkToken struct {
	Purpose   string `json:"purpose"`
	Provider  string `json:"provider"`
	Subject   string `json:"subject"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	ExpiresAt int64  `json:"expires_at"`
}

type oauthProviderConfig struct {
	clientID     string
	clientSecret string
	callbackURL  string
	authURL      string
	tokenURL     string
	scopes       []string
}

type oauthUserInfo struct {
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func oauthConfig(provider, callbackBase string) (oauthProviderConfig, bool) {
	callbackURL := strings.TrimRight(callbackBase, "/") + "/auth/" + provider + "/callback"
	switch provider {
	case "google":
		clientID, clientSecret := os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET")
		if clientID == "" || clientSecret == "" {
			return oauthProviderConfig{}, false
		}
		return oauthProviderConfig{
			clientID:     clientID,
			clientSecret: clientSecret,
			callbackURL:  callbackURL,
			authURL:      "https://accounts.google.com/o/oauth2/v2/auth",
			tokenURL:     "https://oauth2.googleapis.com/token",
			scopes:       []string{"openid", "email", "profile"},
		}, true
	case "github":
		clientID, clientSecret := os.Getenv("GITHUB_OAUTH_CLIENT_ID"), os.Getenv("GITHUB_OAUTH_CLIENT_SECRET")
		if clientID == "" || clientSecret == "" {
			return oauthProviderConfig{}, false
		}
		return oauthProviderConfig{
			clientID:     clientID,
			clientSecret: clientSecret,
			callbackURL:  callbackURL,
			authURL:      "https://github.com/login/oauth/authorize",
			tokenURL:     "https://github.com/login/oauth/access_token",
			scopes:       []string{"read:user", "user:email"},
		}, true
	default:
		return oauthProviderConfig{}, false
	}
}

func (h *Handler) OAuthStart(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	callbackBase := os.Getenv("OAUTH_CALLBACK_BASE_URL")
	if callbackBase == "" {
		http.Error(w, "OAuth is not configured", http.StatusServiceUnavailable)
		return
	}
	providerConfig, ok := oauthConfig(provider, callbackBase)
	if !ok {
		http.Error(w, "OAuth provider is unsupported or not configured", http.StatusServiceUnavailable)
		return
	}
	returnURL, ok := validatedOAuthReturnURL(r.URL.Query().Get("return_to"))
	if !ok {
		http.Error(w, "Invalid OAuth return URL", http.StatusBadRequest)
		return
	}

	mode := r.URL.Query().Get("mode")
	if mode != "" && mode != oauthModeLink {
		http.Error(w, "Invalid OAuth mode", http.StatusBadRequest)
		return
	}

	nonceBytes := make([]byte, 32)
	if _, err := rand.Read(nonceBytes); err != nil {
		http.Error(w, "Unable to initialize OAuth", http.StatusInternalServerError)
		return
	}
	nonce := base64.RawURLEncoding.EncodeToString(nonceBytes)
	state, err := signOAuthState(h.app.Config.JWTSecret, oauthState{
		Nonce: nonce, ReturnURL: returnURL, ExpiresAt: time.Now().Add(5 * time.Minute).Unix(), Mode: mode,
	})
	if err != nil {
		http.Error(w, "Unable to initialize OAuth", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookiePrefix + provider,
		Value:    nonce,
		Path:     "/",
		Domain:   os.Getenv("OAUTH_COOKIE_DOMAIN"),
		HttpOnly: true,
		Secure:   strings.HasPrefix(callbackBase, "https://"),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   300,
	})
	authorizeURL, err := url.Parse(providerConfig.authURL)
	if err != nil {
		http.Error(w, "OAuth provider is misconfigured", http.StatusInternalServerError)
		return
	}
	query := authorizeURL.Query()
	query.Set("client_id", providerConfig.clientID)
	query.Set("redirect_uri", providerConfig.callbackURL)
	query.Set("response_type", "code")
	query.Set("scope", strings.Join(providerConfig.scopes, " "))
	query.Set("state", state)
	authorizeURL.RawQuery = query.Encode()
	http.Redirect(w, r, authorizeURL.String(), http.StatusFound)
}

func (h *Handler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	callbackBase := os.Getenv("OAUTH_CALLBACK_BASE_URL")
	providerConfig, ok := oauthConfig(provider, callbackBase)
	if !ok || callbackBase == "" {
		http.Error(w, "OAuth provider is not configured", http.StatusServiceUnavailable)
		return
	}

	state, err := verifyOAuthState(h.app.Config.JWTSecret, r.URL.Query().Get("state"))
	cookie, cookieErr := r.Cookie(oauthStateCookiePrefix + provider)
	http.SetCookie(w, &http.Cookie{
		Name: oauthStateCookiePrefix + provider, Value: "", Path: "/",
		Domain:   os.Getenv("OAUTH_COOKIE_DOMAIN"),
		HttpOnly: true, Secure: strings.HasPrefix(callbackBase, "https://"),
		SameSite: http.SameSiteLaxMode, MaxAge: -1,
	})
	if err != nil || cookieErr != nil || !hmac.Equal([]byte(cookie.Value), []byte(state.Nonce)) ||
		state.ExpiresAt < time.Now().Unix() {
		if state == nil {
			http.Error(w, "OAuth state validation failed", http.StatusUnauthorized)
			return
		}
		h.redirectOAuthError(w, state.ReturnURL, "invalid_state")
		return
	}
	if providerError := r.URL.Query().Get("error"); providerError != "" {
		h.redirectOAuthError(w, state.ReturnURL, "authorization_denied")
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		h.redirectOAuthError(w, state.ReturnURL, "missing_code")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	identity, err := fetchOAuthIdentity(ctx, provider, providerConfig, code)
	if err != nil {
		h.app.Logger.Warn("OAuth provider identity lookup failed", zap.String("provider", provider), zap.Error(err))
		h.redirectOAuthError(w, state.ReturnURL, "provider_failed")
		return
	}
	identity.Provider = enum.LoginProvider(provider)
	if state.Mode == oauthModeLink {
		linkToken, signErr := signPayload(h.app.Config.JWTSecret, oauthLinkToken{
			Purpose: oauthLinkPurpose, Provider: provider, Subject: identity.Subject,
			Email: identity.Email, Name: identity.Name, AvatarURL: identity.AvatarURL,
			ExpiresAt: time.Now().Add(oauthLinkTokenTTL).Unix(),
		})
		if signErr != nil {
			h.redirectOAuthError(w, state.ReturnURL, "account_failed")
			return
		}
		http.Redirect(w, r, state.ReturnURL+"#oauth_link_token="+url.QueryEscape(linkToken)+
			"&oauth_provider="+url.QueryEscape(provider), http.StatusFound)
		return
	}
	_, err = user_services.LoginWithOAuth(ctx, h.app, identity)
	if err != nil {
		h.app.Logger.Warn("OAuth account login failed", zap.String("provider", provider), zap.Error(err))
		h.redirectOAuthError(w, state.ReturnURL, "account_failed")
		return
	}

	// 登录凭证已通过 HttpOnly 会话 Cookie 下发，URL 里不再携带 token
	redirectURL := state.ReturnURL + "#oauth_login=success&oauth_provider=" + url.QueryEscape(provider)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func validatedOAuthReturnURL(raw string) (string, bool) {
	allowed := allowedOAuthOrigins()
	if raw == "" {
		if len(allowed) == 0 {
			return "", false
		}
		return allowed[0] + "/", true
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.User != nil || parsed.Host == "" ||
		(parsed.Scheme != "https" && parsed.Scheme != "http") {
		return "", false
	}
	origin := parsed.Scheme + "://" + parsed.Host
	isAllowed := false
	for _, candidate := range allowed {
		if origin == candidate {
			isAllowed = true
			break
		}
	}
	if !isAllowed || strings.HasPrefix(parsed.Path, "//") {
		return "", false
	}
	path := parsed.EscapedPath()
	if path == "" {
		path = "/"
	}
	if parsed.RawQuery != "" {
		path += "?" + parsed.RawQuery
	}
	return origin + path, true
}

func allowedOAuthOrigins() []string {
	raw := os.Getenv("OAUTH_ALLOWED_REDIRECT_ORIGINS")
	if raw == "" {
		raw = "https://moelink.site,https://www.moelink.site,https://quiz.moelink.site,http://localhost:3000,http://localhost:5173,http://localhost:8000,https://api.moelink.site"
	}
	var origins []string
	for _, item := range strings.Split(raw, ",") {
		origin := strings.TrimRight(strings.TrimSpace(item), "/")
		if origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}

func signOAuthState(secret string, state oauthState) (string, error) {
	return signPayload(secret, state)
}

func signPayload(secret string, value any) (string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write([]byte(encoded)); err != nil {
		return "", err
	}
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return encoded + "." + signature, nil
}

func verifyOAuthState(secret, token string) (*oauthState, error) {
	var state oauthState
	if err := verifyPayload(secret, token, &state); err != nil {
		return nil, err
	}
	if _, ok := validatedOAuthReturnURL(state.ReturnURL); !ok {
		return nil, fmt.Errorf("invalid return URL")
	}
	return &state, nil
}

func verifyPayload(secret, token string, out any) error {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid token")
	}
	provided, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("invalid token signature")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write([]byte(parts[0])); err != nil {
		return err
	}
	if !hmac.Equal(provided, mac.Sum(nil)) {
		return fmt.Errorf("invalid token signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, out)
}

func verifyOAuthLinkToken(secret, token string) (*oauthLinkToken, error) {
	var link oauthLinkToken
	if err := verifyPayload(secret, token, &link); err != nil {
		return nil, err
	}
	if link.Purpose != oauthLinkPurpose || link.ExpiresAt < time.Now().Unix() ||
		(link.Provider != "google" && link.Provider != "github") || link.Subject == "" {
		return nil, fmt.Errorf("invalid link token")
	}
	return &link, nil
}

func (h *Handler) redirectOAuthError(w http.ResponseWriter, returnURL, code string) {
	if _, ok := validatedOAuthReturnURL(returnURL); !ok {
		http.Error(w, "OAuth login failed", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Location", returnURL+"#oauth_error="+url.QueryEscape(code))
	w.WriteHeader(http.StatusFound)
}

func fetchOAuthIdentity(ctx context.Context, provider string, oauthConfig oauthProviderConfig, code string) (user_services.OAuthIdentity, error) {
	form := url.Values{
		"client_id":     {oauthConfig.clientID},
		"client_secret": {oauthConfig.clientSecret},
		"code":          {code},
		"redirect_uri":  {oauthConfig.callbackURL},
		"grant_type":    {"authorization_code"},
	}
	tokenReq, err := http.NewRequestWithContext(ctx, http.MethodPost, oauthConfig.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return user_services.OAuthIdentity{}, err
	}
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokenReq.Header.Set("Accept", "application/json")
	client := &http.Client{Timeout: 15 * time.Second}
	tokenResp, err := client.Do(tokenReq)
	if err != nil {
		return user_services.OAuthIdentity{}, err
	}
	defer tokenResp.Body.Close()
	if tokenResp.StatusCode != http.StatusOK {
		return user_services.OAuthIdentity{}, fmt.Errorf("OAuth token exchange returned %s", tokenResp.Status)
	}
	var tokenData struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(io.LimitReader(tokenResp.Body, 1<<20)).Decode(&tokenData); err != nil {
		return user_services.OAuthIdentity{}, err
	}
	if tokenData.AccessToken == "" {
		return user_services.OAuthIdentity{}, fmt.Errorf("OAuth token exchange returned no access token")
	}

	if provider == "google" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://openidconnect.googleapis.com/v1/userinfo", nil)
		if err != nil {
			return user_services.OAuthIdentity{}, err
		}
		req.Header.Set("Authorization", "Bearer "+tokenData.AccessToken)
		resp, err := client.Do(req)
		if err != nil {
			return user_services.OAuthIdentity{}, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return user_services.OAuthIdentity{}, fmt.Errorf("Google userinfo returned %s", resp.Status)
		}
		var profile oauthUserInfo
		if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&profile); err != nil {
			return user_services.OAuthIdentity{}, err
		}
		if profile.Subject == "" || !profile.EmailVerified {
			return user_services.OAuthIdentity{}, fmt.Errorf("Google account has no verified email")
		}
		return user_services.OAuthIdentity{
			Subject: profile.Subject, Email: profile.Email, EmailVerified: profile.EmailVerified,
			Name: profile.Name, AvatarURL: profile.Picture,
		}, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return user_services.OAuthIdentity{}, err
	}
	req.Header.Set("Authorization", "Bearer "+tokenData.AccessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := client.Do(req)
	if err != nil {
		return user_services.OAuthIdentity{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return user_services.OAuthIdentity{}, fmt.Errorf("GitHub userinfo returned %s", resp.Status)
	}
	var githubUser struct {
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&githubUser); err != nil {
		return user_services.OAuthIdentity{}, err
	}
	if githubUser.ID == 0 {
		return user_services.OAuthIdentity{}, fmt.Errorf("GitHub returned no user id")
	}

	emailReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return user_services.OAuthIdentity{}, err
	}
	emailReq.Header.Set("Authorization", "Bearer "+tokenData.AccessToken)
	emailReq.Header.Set("Accept", "application/vnd.github+json")
	emailReq.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	emailResp, err := client.Do(emailReq)
	if err != nil {
		return user_services.OAuthIdentity{}, err
	}
	defer emailResp.Body.Close()
	if emailResp.StatusCode != http.StatusOK {
		return user_services.OAuthIdentity{}, fmt.Errorf("GitHub email lookup returned %s", emailResp.Status)
	}
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(io.LimitReader(emailResp.Body, 1<<20)).Decode(&emails); err != nil {
		return user_services.OAuthIdentity{}, err
	}
	email := ""
	for _, candidate := range emails {
		if candidate.Primary && candidate.Verified {
			email = candidate.Email
			break
		}
	}
	if email == "" {
		for _, candidate := range emails {
			if candidate.Verified {
				email = candidate.Email
				break
			}
		}
	}
	if email == "" {
		return user_services.OAuthIdentity{}, fmt.Errorf("GitHub account has no verified email")
	}
	name := githubUser.Name
	if name == "" {
		name = githubUser.Login
	}
	return user_services.OAuthIdentity{
		Subject: fmt.Sprintf("%d", githubUser.ID), Email: email, EmailVerified: true,
		Name: name, AvatarURL: githubUser.AvatarURL,
	}, nil
}
