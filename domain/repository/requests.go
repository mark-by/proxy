package repository

import "github.com/mark-by/proxy/domain/entity"

type Requests interface {
	List() []entity.Request
	Save(rawRequest string) (uint64, error)
	Delete(id uint64) error
	DeleteAll() error
	Get(id uint64) *entity.Request
}
