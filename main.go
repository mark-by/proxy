package main

import (
	"github.com/mark-by/proxy/application"
	"github.com/mark-by/proxy/infrastructure/persistent"
	"github.com/mark-by/proxy/interfaces/proxy"
	"github.com/mark-by/proxy/interfaces/repeater"
)

func init() {
	persistent.Migrate()
}

func main() {
	repositories := persistent.NewRepositories()
	app := application.New(repositories)

	go proxy.Server(app)
	repeater.Server(app)
}
