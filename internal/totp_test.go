package internal

import (
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

func TestNormalizeSecret(t *testing.T) {
	raw := " abcd efgh-ijkl "
	want := "ABCDEFGHIJKL"
	got := normalizeSecret(raw)
	if got != want {
		t.Errorf("normalizeSecret(%q) = %q; want %q", raw, got, want)
	}
}

func TestGetTOTPCode(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP" // base32 "Hello!" example
	account := Account{Secret: secret, Digits: 6, Period: 30, Algorithm: "SHA1"}
	refCode, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatalf("ref totp generate failed: %v", err)
	}
	code, err := getTOTPCode(account)
	if err != nil {
		t.Fatalf("getTOTPCode failed: %v", err)
	}
	if !strings.EqualFold(code, refCode) {
		t.Errorf("getTOTPCode = %q; want %q", code, refCode)
	}
}
