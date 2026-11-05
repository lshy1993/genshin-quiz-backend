package handler

import (
	"os"
	"testing"
	"time"
)

func TestOAuthStateRoundTrip(t *testing.T) {
	t.Setenv("OAUTH_ALLOWED_REDIRECT_ORIGINS", "https://moelink.site")
	state := oauthState{
		Nonce:     "random-state",
		ReturnURL: "https://moelink.site/login",
		ExpiresAt: time.Now().Add(time.Minute).Unix(),
	}

	token, err := signOAuthState("test-secret", state)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := verifyOAuthState("test-secret", token)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Nonce != state.Nonce || decoded.ReturnURL != state.ReturnURL {
		t.Fatalf("decoded state = %#v, want %#v", decoded, state)
	}
}

func TestOAuthStateRejectsTampering(t *testing.T) {
	t.Setenv("OAUTH_ALLOWED_REDIRECT_ORIGINS", "https://moelink.site")
	token, err := signOAuthState("test-secret", oauthState{
		Nonce:     "random-state",
		ReturnURL: "https://moelink.site/",
		ExpiresAt: time.Now().Add(time.Minute).Unix(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := verifyOAuthState("test-secret", token+"x"); err == nil {
		t.Fatal("expected tampered state to be rejected")
	}
}

func TestOAuthReturnURLRequiresAllowedOrigin(t *testing.T) {
	t.Setenv("OAUTH_ALLOWED_REDIRECT_ORIGINS", "https://moelink.site")
	if _, ok := validatedOAuthReturnURL("https://evil.example/"); ok {
		t.Fatal("unexpectedly accepted untrusted return origin")
	}
	if got, ok := validatedOAuthReturnURL("https://moelink.site/login"); !ok || got != "https://moelink.site/login" {
		t.Fatalf("validated OAuth return URL = %q, %t", got, ok)
	}
}

func TestOAuthDefaultOriginsIncludeBothApplications(t *testing.T) {
	if err := os.Unsetenv("OAUTH_ALLOWED_REDIRECT_ORIGINS"); err != nil {
		t.Fatal(err)
	}
	origins := allowedOAuthOrigins()
	foundMoeLink, foundQuiz := false, false
	for _, origin := range origins {
		foundMoeLink = foundMoeLink || origin == "https://moelink.site"
		foundQuiz = foundQuiz || origin == "https://quiz.moelink.site"
	}
	if !foundMoeLink || !foundQuiz {
		t.Fatalf("default origins do not include both sites: %#v", origins)
	}
}

func TestOAuthGoogleCallbackURL(t *testing.T) {
	t.Setenv("GOOGLE_CLIENT_ID", "client-id")
	t.Setenv("GOOGLE_CLIENT_SECRET", "client-secret")
	provider, ok := oauthConfig("google", "https://api.moelink.site")
	if !ok {
		t.Fatal("expected Google OAuth config")
	}
	if provider.callbackURL != "https://api.moelink.site/auth/google/callback" {
		t.Fatalf("Google callback URL = %q", provider.callbackURL)
	}
}

func TestOAuthLinkTokenRoundTripAndRejection(t *testing.T) {
	valid := oauthLinkToken{
		Purpose: oauthLinkPurpose, Provider: "github", Subject: "42",
		ExpiresAt: time.Now().Add(time.Minute).Unix(),
	}
	token, err := signPayload("test-secret", valid)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := verifyOAuthLinkToken("test-secret", token)
	if err != nil || decoded.Subject != "42" || decoded.Provider != "github" {
		t.Fatalf("decoded = %#v, err = %v", decoded, err)
	}
	if _, err := verifyOAuthLinkToken("other-secret", token); err == nil {
		t.Fatal("expected wrong secret to be rejected")
	}

	expired := valid
	expired.ExpiresAt = time.Now().Add(-time.Minute).Unix()
	token, _ = signPayload("test-secret", expired)
	if _, err := verifyOAuthLinkToken("test-secret", token); err == nil {
		t.Fatal("expected expired token to be rejected")
	}

	wrongPurpose := valid
	wrongPurpose.Purpose = "other"
	token, _ = signPayload("test-secret", wrongPurpose)
	if _, err := verifyOAuthLinkToken("test-secret", token); err == nil {
		t.Fatal("expected wrong purpose to be rejected")
	}
}
