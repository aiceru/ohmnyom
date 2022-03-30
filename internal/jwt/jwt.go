package jwt

import (
	"time"

	"github.com/golang-jwt/jwt"
	"ohmnyom/internal/errors"
)

const (
	subjectAuth    = "ohmnyom-auth"
	subjectRefresh = "ohmnyom-refresh"
	tokenIssuer    = "ohmnyom-admin-aiceru"
	expAuth        = time.Hour * 24
	expRefresh     = time.Hour * 24 * 30
)

type Manager struct {
	secret []byte
}

type claim struct {
	Uid string `json:"uid"`
	jwt.StandardClaims
}

func NewManager(secret []byte) *Manager {
	if len(secret) < 1 {
		return nil
	}
	return &Manager{secret: secret}
}

func (m *Manager) newToken(uid string, expDuration time.Duration, subject string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim{
		Uid: uid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expDuration).UTC().Unix(),
			IssuedAt:  time.Now().UTC().Unix(),
			Issuer:    tokenIssuer,
			Subject:   subject,
		},
	})
	tokenString, err := token.SignedString(m.secret)
	if err != nil {
		return "", errors.New("%v", err)
	}
	return tokenString, nil
}

func (m *Manager) NewAuthToken(uid string) (string, error) {
	return m.newToken(uid, expAuth, subjectAuth)
}

func (m *Manager) NewRefreshToken(uid string) (string, error) {
	return m.newToken(uid, expRefresh, subjectRefresh)
}

func (m *Manager) Verify(tokenStr string) (string, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&claim{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, errors.NewAuthenticationError("jwt invalid signing method")
			}
			return m.secret, nil
		},
	)
	if err != nil {
		return "", errors.NewAuthenticationError(err.Error())
	}

	c, ok := token.Claims.(*claim)
	if !ok {
		return "", errors.NewAuthenticationError("invalid claim")
	}

	return c.Uid, nil
}
