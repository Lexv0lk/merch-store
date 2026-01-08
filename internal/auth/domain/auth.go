package domain

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenIssuer interface {
	IssueToken(secret []byte, userID int, username string, timeLimit time.Duration) (string, error)
}

type Claims struct {
	UserID   int64  `json:"uid"`
	Username string `json:"usr"`
	jwt.RegisteredClaims
}

type JWTTokenIssuer struct {
}

func NewJWTTokenIssuer() *JWTTokenIssuer {
	return &JWTTokenIssuer{}
}

func (ti *JWTTokenIssuer) IssueToken(secret []byte, userID int, username string, timeLimit time.Duration) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID:   int64(userID),
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(int64(userID), 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(timeLimit)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}
