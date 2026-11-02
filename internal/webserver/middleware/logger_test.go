package middleware

import (
	"net/http/httptest"
	"testing"
)

func TestSafeLogURLRedactsOAuthQuery(t *testing.T) {
	req := httptest.NewRequest("GET", "/auth/google/callback?code=secret-code&state=secret-state", nil)
	if got := safeLogURL(req); got != "/auth/google/callback" {
		t.Fatalf("safeLogURL() = %q, want path without OAuth query", got)
	}
}

func TestSafeLogURLKeepsOtherQuery(t *testing.T) {
	req := httptest.NewRequest("GET", "/health?probe=ready", nil)
	if got := safeLogURL(req); got != "/health?probe=ready" {
		t.Fatalf("safeLogURL() = %q, want original URL", got)
	}
}
