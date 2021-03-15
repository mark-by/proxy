package repeater

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/mark-by/proxy/application"
	"net/http"
	"strconv"
)

func listRequests(writer http.ResponseWriter, request *http.Request, app *application.App) {
	requests := app.Requests.GetAll()
	if len(requests) == 0 {
		writer.WriteHeader(http.StatusNoContent)
		return
	}

	data, err := json.Marshal(requests)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Add("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getRequest(writer http.ResponseWriter, request *http.Request, app *application.App) {
	vars := mux.Vars(request)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	r := app.Requests.Get(id)
	if r == nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	data, err := json.Marshal(r)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Add("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func deleteRequest(writer http.ResponseWriter, request *http.Request, app *application.App) {
	vars := mux.Vars(request)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	err = app.Requests.Delete(id)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
}

func deleteRequests(writer http.ResponseWriter, request *http.Request, app *application.App) {
	err := app.Requests.DeleteAll()
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func repeatRequest(writer http.ResponseWriter, request *http.Request, app *application.App) {
	vars := mux.Vars(request)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	r := app.Requests.Get(id)
	if r == nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	defer client.CloseIdleConnections()

	liveRequest, err := r.Revive()
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := client.Do(liveRequest)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	application.CopyResponse(response, writer)
}

func scanCommandInjection(writer http.ResponseWriter, request *http.Request, app *application.App) {

}
