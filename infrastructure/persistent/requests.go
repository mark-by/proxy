package persistent

import (
	"github.com/jackc/pgx"
	"github.com/mark-by/proxy/domain/entity"
	"github.com/mark-by/proxy/domain/repository"
)

type Request struct {
	db *pgx.ConnPool
}

func newRequestRepo(db *pgx.ConnPool) *Request {
	return &Request{db}
}

func (r Request) List() []entity.Request {
	panic("implement me")
}

func (r Request) Save(rawRequest string) (uint64, error) {
	panic("implement me")
}

func (r Request) Delete(id uint64) error {
	panic("implement me")
}

func (r Request) DeleteAll() error {
	panic("implement me")
}

func (r Request) Get(id uint64) *entity.Request {
	panic("implement me")
}

var _ repository.Requests = &Request{}
