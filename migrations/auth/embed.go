package authmigrationsfs

import "embed"

//go:embed *.sql
var AuthMigrations embed.FS
