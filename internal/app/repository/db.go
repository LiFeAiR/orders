package repository

import (
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewORM(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(gormPostgres.New(gormPostgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
