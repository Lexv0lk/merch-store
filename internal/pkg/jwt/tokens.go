package jwt

import (
	"context"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Authenticator interface {
	Authenticate(ctx context.Context, username, password string) (string, error)
}

type TokenIssuer interface {
	IssueToken(secret []byte, userID int, username string, timeLimit time.Duration) (string, error)
}

type Claims struct {
	UserID   int    `json:"uid"`
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
		UserID:   userID,
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

type JWTTokenParser struct {
}

func NewJWTTokenParser() *JWTTokenParser {
	return &JWTTokenParser{}
}

func (tp *JWTTokenParser) ParseToken(secret []byte, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenUnverifiable
		}

		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}
