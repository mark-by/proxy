package application

import "net/http"

type IRequests interface {
	Intercept(w http.ResponseWriter, r *http.Request)
}
