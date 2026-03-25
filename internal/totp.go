package internal

import (
	"fmt"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// Vault is a container for many accounts, saved encrypted
type Vault struct {
	Accounts []Account `json:"accounts"`
}

// Account holds one 2FA account definition.
type Account struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Issuer    string `json:"issuer"`
	Secret    string `json:"secret"`
	Algorithm string `json:"algorithm"`
	Digits    int    `json:"digits"`
	Period    int    `json:"period"`
}

func normalizeSecret(raw string) string {
	clean := strings.ToUpper(strings.ReplaceAll(raw, " ", ""))
	clean = strings.ReplaceAll(clean, "-", "")
	clean = strings.TrimSpace(clean)
	return clean
}

func getTOTPCode(account Account) (string, error) {
	secret := normalizeSecret(account.Secret)
	if account.Digits <= 0 {
		account.Digits = 6
	}
	period := account.Period
	if period <= 0 {
		period = 30
	}
	algo := otp.AlgorithmSHA1
	if strings.EqualFold(account.Algorithm, "SHA256") {
		algo = otp.AlgorithmSHA256
	} else if strings.EqualFold(account.Algorithm, "SHA512") {
		algo = otp.AlgorithmSHA512
	}
	code, err := totp.GenerateCodeCustom(secret, time.Now(), totp.ValidateOpts{
		Period:    uint(period),
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: algo,
	})
	if err != nil {
		return "", fmt.Errorf("unable to generate TOTP: %w", err)
	}
	return code, nil
}
