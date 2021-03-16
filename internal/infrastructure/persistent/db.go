package persistent

import (
	"fmt"
	"github.com/jackc/pgx"
	"github.com/mark-by/proxy/internal/config"
	"github.com/mark-by/proxy/internal/domain/repository"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func _migrate() error {
	m, err := migrate.New(
		"file://internal/infrastructure/persistent/migrations",
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			viper.GetString(config.DBUser),
			viper.GetString(config.DBPassword),
			viper.GetString(config.DBHost),
			viper.GetString(config.DBPort),
			viper.GetString(config.DBName),
		),
	)
	if err != nil {
		logrus.Error("Fail to connect to database: ", err)
		return err
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
	return nil
}

func Migrate() {
	for idx := 0; idx < 3; idx++ {
		err := _migrate()
		if err != nil {
			time.Sleep(5 * time.Second)
		} else {
			return
		}
	}

	logrus.Fatal("Fail to migrate")
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
