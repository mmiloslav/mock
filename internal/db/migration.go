package db

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var migrations = []*gormigrate.Migration{
	{
		ID:      "migrate_20250521_initial",
		Migrate: migrate_20250521_initial,
	},
}

func migrate_20250521_initial(tx *gorm.DB) error {
	return tx.AutoMigrate(
		&Mock{},
		&Group{},
	)
}
