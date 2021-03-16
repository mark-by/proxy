package main

import (
	"github.com/mark-by/proxy/internal/application"
	"github.com/mark-by/proxy/internal/infrastructure/persistent"
	"github.com/mark-by/proxy/internal/interfaces/proxy"
	"github.com/mark-by/proxy/internal/interfaces/repeater"
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
