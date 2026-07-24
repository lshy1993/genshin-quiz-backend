package middleware

import "testing"

func TestIsPublicEndpointHomeRequiresNoAuth(t *testing.T) {
	if !isPublicEndpoint("/home", "GET") {
		t.Fatalf("expected /home GET to be treated as a public endpoint")
	}
}
