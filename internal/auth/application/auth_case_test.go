//go:generate mockgen
package application

import (
	"testing"
	"time"

	authmocks "github.com/Lexv0lk/merch-store/gen/mocks/auth"
	jwtmocks "github.com/Lexv0lk/merch-store/gen/mocks/jwt"
	"github.com/Lexv0lk/merch-store/internal/auth/domain"
	"github.com/Lexv0lk/merch-store/internal/pkg/jwt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticator_Authenticate(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name               string
		username, password string
		secretKey          string

		prepareFn func(t *testing.T, ctrl *gomock.Controller) (domain.UsersRepository, domain.PasswordHasher, jwt.TokenIssuer)

		expectedToken string
		expectedErr   error
	}

	tests := []testCase{
		{
			name:      "new user created successfully",
			username:  "newuser",
			password:  "password123",
			secretKey: "secret",
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) (domain.UsersRepository, domain.PasswordHasher, jwt.TokenIssuer) {
				usersRepo := authmocks.NewMockUsersRepository(ctrl)
				passwordHasher := authmocks.NewMockPasswordHasher(ctrl)
				tokenIssuer := jwtmocks.NewMockTokenIssuer(ctrl)

				usersRepo.EXPECT().TryGetUserInfo(gomock.Any(), "newuser").Return(domain.UserInfo{}, false, nil)
				passwordHasher.EXPECT().HashPassword("password123").Return("hashed_password", nil)
				usersRepo.EXPECT().CreateUser(gomock.Any(), "newuser", "hashed_password").Return(domain.UserInfo{
					ID:           1,
					Username:     "newuser",
					PasswordHash: "hashed_password",
				}, nil)
				tokenIssuer.EXPECT().IssueToken([]byte("secret"), 1, "newuser", time.Hour).Return("jwt_token", nil)

				return usersRepo, passwordHasher, tokenIssuer
			},
			expectedToken: "jwt_token",
			expectedErr:   nil,
		},
		{
			name:      "existing user with correct password",
			username:  "existinguser",
			password:  "correctpassword",
			secretKey: "secret",
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) (domain.UsersRepository, domain.PasswordHasher, jwt.TokenIssuer) {
				usersRepo := authmocks.NewMockUsersRepository(ctrl)
				passwordHasher := authmocks.NewMockPasswordHasher(ctrl)
				tokenIssuer := jwtmocks.NewMockTokenIssuer(ctrl)

				usersRepo.EXPECT().TryGetUserInfo(gomock.Any(), "existinguser").Return(domain.UserInfo{
					ID:           2,
					Username:     "existinguser",
					PasswordHash: "stored_hash",
				}, true, nil)
				passwordHasher.EXPECT().VerifyPassword("correctpassword", "stored_hash").Return(true, nil)
				tokenIssuer.EXPECT().IssueToken([]byte("secret"), 2, "existinguser", time.Hour).Return("jwt_token", nil)

				return usersRepo, passwordHasher, tokenIssuer
			},
			expectedToken: "jwt_token",
			expectedErr:   nil,
		},
		{
			name:      "existing user with incorrect password",
			username:  "existinguser",
			password:  "wrongpassword",
			secretKey: "secret",
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) (domain.UsersRepository, domain.PasswordHasher, jwt.TokenIssuer) {
				usersRepo := authmocks.NewMockUsersRepository(ctrl)
				passwordHasher := authmocks.NewMockPasswordHasher(ctrl)
				tokenIssuer := jwtmocks.NewMockTokenIssuer(ctrl)

				usersRepo.EXPECT().TryGetUserInfo(gomock.Any(), "existinguser").Return(domain.UserInfo{
					ID:           2,
					Username:     "existinguser",
					PasswordHash: "stored_hash",
				}, true, nil)
				passwordHasher.EXPECT().VerifyPassword("wrongpassword", "stored_hash").Return(false, nil)

				return usersRepo, passwordHasher, tokenIssuer
			},
			expectedToken: "",
			expectedErr:   &domain.CredentialsMismatchError{},
		},
		{
			name:      "error getting user info",
			username:  "testuser",
			password:  "password",
			secretKey: "secret",
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) (domain.UsersRepository, domain.PasswordHasher, jwt.TokenIssuer) {
				usersRepo := authmocks.NewMockUsersRepository(ctrl)
				passwordHasher := authmocks.NewMockPasswordHasher(ctrl)
				tokenIssuer := jwtmocks.NewMockTokenIssuer(ctrl)

				usersRepo.EXPECT().TryGetUserInfo(gomock.Any(), "testuser").Return(domain.UserInfo{}, false, assert.AnError)

				return usersRepo, passwordHasher, tokenIssuer
			},
			expectedToken: "",
			expectedErr:   assert.AnError,
		},
		{
			name:      "error hashing password for new user",
			username:  "newuser",
			password:  "password",
			secretKey: "secret",
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) (domain.UsersRepository, domain.PasswordHasher, jwt.TokenIssuer) {
				usersRepo := authmocks.NewMockUsersRepository(ctrl)
				passwordHasher := authmocks.NewMockPasswordHasher(ctrl)
				tokenIssuer := jwtmocks.NewMockTokenIssuer(ctrl)

				usersRepo.EXPECT().TryGetUserInfo(gomock.Any(), "newuser").Return(domain.UserInfo{}, false, nil)
				passwordHasher.EXPECT().HashPassword("password").Return("", assert.AnError)

				return usersRepo, passwordHasher, tokenIssuer
			},
			expectedToken: "",
			expectedErr:   assert.AnError,
		},
		{
			name:      "error creating new user",
			username:  "newuser",
			password:  "password",
			secretKey: "secret",
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) (domain.UsersRepository, domain.PasswordHasher, jwt.TokenIssuer) {
				usersRepo := authmocks.NewMockUsersRepository(ctrl)
				passwordHasher := authmocks.NewMockPasswordHasher(ctrl)
				tokenIssuer := jwtmocks.NewMockTokenIssuer(ctrl)

				usersRepo.EXPECT().TryGetUserInfo(gomock.Any(), "newuser").Return(domain.UserInfo{}, false, nil)
				passwordHasher.EXPECT().HashPassword("password").Return("hashed_password", nil)
				usersRepo.EXPECT().CreateUser(gomock.Any(), "newuser", "hashed_password").Return(domain.UserInfo{}, assert.AnError)

				return usersRepo, passwordHasher, tokenIssuer
			},
			expectedToken: "",
			expectedErr:   assert.AnError,
		},
		{
			name:      "error verifying password",
			username:  "existinguser",
			password:  "password",
			secretKey: "secret",
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) (domain.UsersRepository, domain.PasswordHasher, jwt.TokenIssuer) {
				usersRepo := authmocks.NewMockUsersRepository(ctrl)
				passwordHasher := authmocks.NewMockPasswordHasher(ctrl)
				tokenIssuer := jwtmocks.NewMockTokenIssuer(ctrl)

				usersRepo.EXPECT().TryGetUserInfo(gomock.Any(), "existinguser").Return(domain.UserInfo{
					ID:           1,
					Username:     "existinguser",
					PasswordHash: "stored_hash",
				}, true, nil)
				passwordHasher.EXPECT().VerifyPassword("password", "stored_hash").Return(false, assert.AnError)

				return usersRepo, passwordHasher, tokenIssuer
			},
			expectedToken: "",
			expectedErr:   assert.AnError,
		},
		{
			name:      "error issuing token for new user",
			username:  "newuser",
			password:  "password",
			secretKey: "secret",
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) (domain.UsersRepository, domain.PasswordHasher, jwt.TokenIssuer) {
				usersRepo := authmocks.NewMockUsersRepository(ctrl)
				passwordHasher := authmocks.NewMockPasswordHasher(ctrl)
				tokenIssuer := jwtmocks.NewMockTokenIssuer(ctrl)

				usersRepo.EXPECT().TryGetUserInfo(gomock.Any(), "newuser").Return(domain.UserInfo{}, false, nil)
				passwordHasher.EXPECT().HashPassword("password").Return("hashed_password", nil)
				usersRepo.EXPECT().CreateUser(gomock.Any(), "newuser", "hashed_password").Return(domain.UserInfo{
					ID:           1,
					Username:     "newuser",
					PasswordHash: "hashed_password",
				}, nil)
				tokenIssuer.EXPECT().IssueToken([]byte("secret"), 1, "newuser", time.Hour).Return("", assert.AnError)

				return usersRepo, passwordHasher, tokenIssuer
			},
			expectedToken: "",
			expectedErr:   assert.AnError,
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			usersRepoMock, passwordHasherMock, tokenIssuerMock := tc.prepareFn(t, ctrl)
			authenticator := NewAuthenticator(usersRepoMock, passwordHasherMock, tokenIssuerMock, tc.secretKey)

			token, err := authenticator.Authenticate(t.Context(), tc.username, tc.password)

			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedToken, token)
			}
		})
	}
}
