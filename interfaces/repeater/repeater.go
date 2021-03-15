package repeater

import (
	"crypto/tls"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mark-by/proxy/application"
	"github.com/mark-by/proxy/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
)



func Server(app *application.App) {
	wrapper := func(handler func (writer http.ResponseWriter, request *http.Request, app *application.App)) func (writer http.ResponseWriter, request *http.Request){
		return func(writer http.ResponseWriter, request *http.Request) {
			handler(writer, request, app)
		}
	}

	r := mux.NewRouter()

	r.HandleFunc("/requests", wrapper(listRequests)).Methods(http.MethodGet)
	r.HandleFunc("/requests", wrapper(deleteRequests)).Methods(http.MethodDelete)

	r.HandleFunc("/requests/{id}", func(writer http.ResponseWriter, request *http.Request) {
		getRequest(writer, request, app)
	}).Methods(http.MethodGet)

	r.HandleFunc("/requests/{id}", wrapper(deleteRequest)).Methods(http.MethodDelete)
	r.HandleFunc("/requests/{id}/repeat", wrapper(repeatRequest)).Methods(http.MethodPost)
	r.HandleFunc("/requests/{id}/scan/cmd", wrapper(scan)).Methods(http.MethodPost)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", viper.Get(config.RepeaterIP), viper.Get(config.RepeaterPort)),
		Handler:      r,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	err := server.ListenAndServe()

	if err != nil {
		logrus.Fatalf("Сервер не запустился с ошибкой: %s", err)
	}
}
