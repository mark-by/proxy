package persistent

import (
	"eco4.ru/work/telephony/config"
	"fmt"
	"github.com/jackc/pgx"
	"github.com/mark-by/proxy/domain/repository"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Migrate() {
	m, err := migrate.New(
		"file://infrastructure/persistent/migrations",
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			viper.GetString(config.DBUser),
			viper.GetString(config.DBPassword),
			viper.GetString(config.DBHost),
			viper.GetString(config.DBPort),
			viper.GetString(config.DBName),
		),
	)
	if err != nil {
		logrus.Fatal("Fail to connect to database: ", err)
	}
	defer m.Close()

	err = m.Up()
	switch err {
	case nil:
		logrus.Info("Migrate status: migrations applied")
	case migrate.ErrNoChange:
		logrus.Info("Migrate status: no changes")
	default:
		logrus.Fatal("Fail to apply migrations: ", err)
	}
}

func NewRepositories() *repository.Repositories {
	conn, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     viper.GetString(config.DBHost),
			Port:     uint16(viper.GetInt(config.DBPort)),
			Database: viper.GetString(config.DBName),
			User:     viper.GetString(config.DBUser),
			Password: viper.GetString(config.DBPassword),
		},
		MaxConnections: 100,
	})

	if err != nil {
		logrus.Fatal("Fail to create db repository: ", err)
	}

	return &repository.Repositories{
		Requests: newRequestRepo(conn),
	}
}
