package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostgresSettings_GetUrl(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		settings    PostgresSettings
		expectedStr string
	}

	tests := []testCase{
		{
			name: "SSL enabled",
			settings: PostgresSettings{
				User:       "testuser",
				Password:   "testpass",
				Host:       "localhost",
				Port:       "5432",
				DBName:     "testdb",
				SSlEnabled: true,
			},
			expectedStr: "postgres://testuser:testpass@localhost:5432/testdb",
		},
		{
			name: "SSL disabled",
			settings: PostgresSettings{
				User:       "testuser",
				Password:   "testpass",
				Host:       "localhost",
				Port:       "5432",
				DBName:     "testdb",
				SSlEnabled: false,
			},
			expectedStr: "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable",
		},
	}

	for _, tc := range tests {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.settings.GetUrl()
			assert.Equal(t, tt.expectedStr, result)
		})
	}
}
