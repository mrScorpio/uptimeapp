package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Tokens struct {
	Secret                []byte
	AccessTTL, RefreshTTL time.Duration
}

func (t Tokens) Access(userID uint) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: strconv.FormatUint(uint64(userID), 10), IssuedAt: jwt.NewNumericDate(time.Now()), ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.AccessTTL))}).SignedString(t.Secret)
}

func (t Tokens) Verify(raw string) (uint, error) {
	token, err := jwt.ParseWithClaims(raw, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("invalid signing method")
		}
		return t.Secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		return 0, errors.New("invalid access token")
	}
	id, err := strconv.ParseUint(token.Claims.(*jwt.RegisteredClaims).Subject, 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("invalid access token")
	}
	return uint(id), nil
}

func Refresh() (string, string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	raw := base64.RawURLEncoding.EncodeToString(b)
	return raw, Hash(raw), nil
}

func Hash(raw string) string { sum := sha256.Sum256([]byte(raw)); return hex.EncodeToString(sum[:]) }
