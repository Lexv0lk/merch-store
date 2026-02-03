package storemigrationsfs

import "embed"

//go:embed *.sql
var StoreMigrations embed.FS
