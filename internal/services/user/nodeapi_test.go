package services

import (
	"strings"
	"testing"
)

func TestPrivacyParams(t *testing.T) {
	public, private, invalid := "public", "private", "friends"

	params, err := ProfileUpdateRequest{EmailVisibility: &public, GenderVisibility: &private}.privacyParams()
	if err != nil {
		t.Fatal(err)
	}
	if *params.EmailVisibility != 1 || *params.GenderVisibility != 0 || params.CountryVisibility != nil {
		t.Fatalf("unexpected params: %+v", params)
	}
	if _, err := (ProfileUpdateRequest{CountryVisibility: &invalid}).privacyParams(); err == nil {
		t.Fatal("expected invalid visibility to be rejected")
	}
	if (ProfileUpdateRequest{}).hasPrivacyChange() {
		t.Fatal("empty request must not report a privacy change")
	}
}

func TestValidateProfileRequest(t *testing.T) {
	str := func(s string) *string { return &s }

	ok := ProfileUpdateRequest{
		Nickname: str("  Traveler "),
		Language: str("zh-CN"), Country: str("jp"),
	}
	if err := validateProfileRequest(&ok); err != nil {
		t.Fatal(err)
	}
	if *ok.Nickname != "Traveler" || *ok.Country != "JP" {
		t.Fatalf("not normalized: %q %q", *ok.Nickname, *ok.Country)
	}

	bad := []ProfileUpdateRequest{
		{Nickname: str("   ")},
		{Nickname: str(strings.Repeat("a", 101))},
		{Bio: str(strings.Repeat("好", 501))},
		{Language: str("not a lang")},
		{Country: str("CHN")},
	}
	for i := range bad {
		if err := validateProfileRequest(&bad[i]); err == nil {
			t.Errorf("case %d: expected error", i)
		}
	}
}
