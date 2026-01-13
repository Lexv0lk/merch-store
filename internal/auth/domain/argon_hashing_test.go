package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArgonHasher(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		password string
	}

	testCases := []testCase{
		{name: "simple password", password: "password123"},
		{name: "complex password", password: "P@ssw0rd!#2024"},
		{name: "empty password", password: ""},
		{name: "long password", password: "aVeryLongPasswordThatExceedsNormalLengthToTestTheHasherFunctionalityAndPerformance1234567890"},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			hasher := NewArgonPasswordHasher()

			hashedPassword, err := hasher.HashPassword(tt.password)
			require.NoError(t, err)

			isValid, err := hasher.VerifyPassword(tt.password, hashedPassword)
			require.NoError(t, err)
			assert.True(t, isValid)
		})
	}
}
