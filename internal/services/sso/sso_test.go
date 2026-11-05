package sso

import (
	"testing"
)

func TestClientsParsing(t *testing.T) {
	t.Setenv("SSO_CLIENTS", `[
		{"id":"partner","name":"Partner","redirect_uris":["https://app.example/cb"]},
		{"id":"","redirect_uris":["https://x.example/cb"]},
		{"id":"no-uris","redirect_uris":[]}
	]`)
	clients := Clients()
	if len(clients) != 1 || clients[0].ID != "partner" {
		t.Fatalf("unexpected clients: %+v", clients)
	}
	if origins := ClientOrigins(); len(origins) != 1 || origins[0] != "https://app.example" {
		t.Fatalf("unexpected origins: %v", origins)
	}

	t.Setenv("SSO_CLIENTS", "not json")
	if Clients() != nil {
		t.Fatal("invalid config must yield no clients")
	}
	t.Setenv("SSO_CLIENTS", "")
	if _, ok := FindClient("partner"); ok {
		t.Fatal("no clients configured, none may be found")
	}
}

func TestRedirectMustMatchExactly(t *testing.T) {
	c := &Client{ID: "p", RedirectURIs: []string{"https://app.example/cb"}}
	for _, uri := range []string{
		"https://app.example/cb",
	} {
		if !c.allowsRedirect(uri) {
			t.Errorf("%q must be allowed", uri)
		}
	}
	for _, uri := range []string{
		"https://app.example/cb/",
		"https://app.example/cb?x=1",
		"https://app.example/cbx",
		"https://app.example.evil.com/cb",
		"http://app.example/cb",
		"",
	} {
		if c.allowsRedirect(uri) {
			t.Errorf("%q must be rejected", uri)
		}
	}
}
