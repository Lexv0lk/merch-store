package database

import (
	"database/sql"
	"io/fs"

	"github.com/pressly/goose/v3"
)

func MigrateDatabase(databaseUrl string, migrations fs.FS, dir, driverName, dialect string) error {
	db, err := sql.Open(driverName, databaseUrl)
	if err != nil {
		return err
	}
	defer db.Close()

	goose.SetBaseFS(migrations)

	if err := goose.SetDialect(dialect); err != nil {
		return err
	}

	if err := goose.Up(db, dir); err != nil {
		return err
	}

	return nil
}
