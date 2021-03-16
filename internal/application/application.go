package application

import "github.com/mark-by/proxy/internal/domain/repository"

type App struct {
	Requests *Requests
}

func New(repositories *repository.Repositories) *App {
	return &App{Requests: newRequests(repositories)}
}
