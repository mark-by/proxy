package config

import (
	"fmt"
	"github.com/mark-by/enviper"
	"github.com/spf13/viper"
	"log"
)

const (
	ProxyIP   = "server.ip"
	ProxyPort = "server.port"

	RepeaterIP   = "repeater.ip"
	RepeaterPort = "repeater.port"

	DBName     = "db.name"
	DBHost     = "db.host"
	DBPort     = "db.port"
	DBUser     = "db.user"
	DBPassword = "db.password"

	Path = "config.yaml"
)

func initConfig() {
	viper.SetConfigFile(Path)

	viperDefaults()
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigParseError); ok {
			log.Fatalf("Невалидный синтаксис")
		} else {
			_ = viper.WriteConfig()
			fmt.Printf("Файл не найден. По пути '%s' записан шаблон\n", Path)
		}
	}

	viper.AllKeys()
}

func init() {
	initConfig()
}

func viperDefaults() {
	hostDefaults()
	dbDefaults()
}

func hostDefaults() {
	// где запускаемся
	viper.SetDefault(ProxyIP, "127.0.0.1")
	viper.SetDefault(ProxyPort, "8080")

	viper.SetDefault(RepeaterIP, "127.0.0.1")
	viper.SetDefault(RepeaterPort, "8888")
}

func dbDefaults() {
	// конфиги бд
	enviper.SetDefaultString(DBName, "")
	enviper.SetDefaultString(DBHost, "127.0.0.1")
	enviper.SetDefaultString(DBPort, "5432")
	enviper.SetDefaultString(DBUser, "")
	enviper.SetDefaultString(DBPassword, "")
}
