package bootstrap

import "github.com/Lexv0lk/merch-store/internal/pkg/database"

type StoreConfig struct {
	DbSettings database.PostgresSettings
	JwtSecret  string
}
