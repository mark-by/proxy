package application

import "github.com/mark-by/proxy/domain/repository"

type App struct {
	Requests *Requests
}

func New(repositories *repository.Repositories) *App {
	return &App{Requests: newRequests(repositories)}
}
