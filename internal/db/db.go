package db

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/mmiloslav/mock/pkg/env"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var mockDB *gorm.DB

func Ping() error {
	sqlDB, err := mockDB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Ping()
}

// Connection Params for DB struct
type ConnectionParams struct {
	User     string
	Password string
	Host     string
	Port     string
	DBName   string
}

// OpenConnection opens db connection
func OpenConnection(logger *logrus.Entry) error {
	cp, err := getConnectionParams()
	if err != nil {
		logger.Errorf("failed to get connection params with error [%s]", err.Error())
		return err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", cp.User, cp.Password, cp.Host, cp.Port)
	d, err := gorm.Open(mysql.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		logger.Errorf("failed to open database with error [%s]", err.Error())
		return err
	}

	err = d.Exec("CREATE DATABASE IF NOT EXISTS `" + cp.DBName + "`;").Error
	if err != nil {
		logger.Errorf("failed create database with error [%s]", err.Error())
		return err
	}

	dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4", cp.User, cp.Password, cp.Host, cp.Port, cp.DBName)
	mockDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		logger.Errorf("failed to open database with error [%s]", err.Error())
		return err
	}

	logger.Info("migrating tables...")
	migrator := gormigrate.New(mockDB, gormigrate.DefaultOptions, migrations)
	err = migrator.Migrate()
	if err != nil {
		logger.Errorf("failed to migrate tables with error [%s]", err.Error())
		return err
	}
	logger.Info("successfully migrated migrations")

	mockDB = mockDB.Debug()

	return nil
}

// getConnectionParams gets DB connection params from env
func getConnectionParams() (ConnectionParams, error) {
	user, err := env.GetVar("MYSQL_USER")
	if err != nil {
		return ConnectionParams{}, err
	}

	password, err := env.GetVar("MYSQL_PASSWORD")
	if err != nil {
		return ConnectionParams{}, err
	}

	host, err := env.GetVar("MYSQL_HOST")
	if err != nil {
		return ConnectionParams{}, err
	}

	port, err := env.GetVar("MYSQL_PORT")
	if err != nil {
		return ConnectionParams{}, err
	}

	dbName, err := env.GetVar("MYSQL_DATABASE")
	if err != nil {
		return ConnectionParams{}, err
	}

	return ConnectionParams{
		User:     user,
		Password: password,
		Host:     host,
		Port:     port,
		DBName:   dbName,
	}, nil
}
