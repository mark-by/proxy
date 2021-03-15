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

func (r Request) List() ([]entity.Request, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {endTx(tx, err)}()

	rows, err := tx.Query("select id, raw, url from requests")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []entity.Request
	for rows.Next() {
		var request entity.Request
		err = rows.Scan(&request.ID, &request.Raw, &request.URL)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}

	return requests, nil
}

func (r Request) Save(url, rawRequest string) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {endTx(tx, err)}()
	var ID int
	err = tx.QueryRow("insert into requests (raw, url) values ($1, $2) returning id", rawRequest, url).Scan(&ID)
	if err != nil {
		return 0, nil
	}
	return ID, nil
}

func (r Request) Delete(id int) error {
	panic("implement me")
}

func (r Request) DeleteAll() error {
	panic("implement me")
}

func (r Request) Get(id int) (*entity.Request, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {endTx(tx, err)}()

	var request entity.Request
	err = tx.QueryRow("select raw, url from requests where id = $1", id).Scan(&request.Raw, &request.URL)
	if err != nil {
		return nil, err
	}

	request.ID = id
	return &request, nil
}

var _ repository.Requests = &Request{}
