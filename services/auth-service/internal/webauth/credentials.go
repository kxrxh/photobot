package webauth

import (
	"crypto/rand"
	"errors"
	"math/big"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost          = 12
	recoveryCodeCount   = 8
	recoveryCodeCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
)

var loginPattern = regexp.MustCompile(`^[a-z0-9_-]{3,32}$`)

func NormalizeLogin(login string) string {
	return strings.ToLower(strings.TrimSpace(login))
}

func ValidateLoginFormat(login string) bool {
	return loginPattern.MatchString(login)
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func GenerateRecoveryCodes() (plain []string, hashes []string, err error) {
	plain = make([]string, 0, recoveryCodeCount)
	hashes = make([]string, 0, recoveryCodeCount)
	for range recoveryCodeCount {
		code, genErr := randomRecoveryCode()
		if genErr != nil {
			return nil, nil, genErr
		}
		hash, hashErr := HashPassword(code)
		if hashErr != nil {
			return nil, nil, hashErr
		}
		plain = append(plain, code)
		hashes = append(hashes, hash)
	}
	return plain, hashes, nil
}

func randomRecoveryCode() (string, error) {
	const partLen = 4
	var b strings.Builder
	for i := range 2 {
		if i > 0 {
			b.WriteByte('-')
		}
		for range partLen {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(len(recoveryCodeCharset))))
			if err != nil {
				return "", err
			}
			b.WriteByte(recoveryCodeCharset[n.Int64()])
		}
	}
	return b.String(), nil
}

func NormalizeRecoveryCode(code string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(code), " ", ""))
}

func ValidatePassword(password string) error {
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	return nil
}
