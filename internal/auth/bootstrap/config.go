package bootstrap

import "github.com/Lexv0lk/merch-store/internal/pkg/database"

type AuthConfig struct {
	DbSettings database.PostgresSettings
	SecretKey  string
}
