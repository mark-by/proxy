package application

import (
	"github.com/mark-by/proxy/domain/entity"
	"net/http"
)

type IRequests interface {
	Intercept(w http.ResponseWriter, r *http.Request)
	Get(id int) *entity.Request
	GetAll() []entity.Request
}
