package repository

import "github.com/mark-by/proxy/internal/domain/entity"

type Requests interface {
	List() ([]entity.Request, error)
	Save(url, rawRequest string) (int, error)
	Delete(id int) error
	DeleteAll() error
	Get(id int) (*entity.Request, error)
}
