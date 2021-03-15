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
	r := mux.NewRouter()

	r.HandleFunc("requests", listRequests).Methods(http.MethodGet)
	r.HandleFunc("requests", deleteRequests).Methods(http.MethodDelete)
	r.HandleFunc("requests/{id}", getRequest).Methods(http.MethodGet)
	r.HandleFunc("requests/{id}", deleteRequest).Methods(http.MethodDelete)
	r.HandleFunc("requests/{id}/repeat", repeatRequest).Methods(http.MethodPost)
	r.HandleFunc("requests/{id}/commandInjectionScan", scanCommandInjection).Methods(http.MethodPost)

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
