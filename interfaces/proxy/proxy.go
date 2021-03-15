package proxy

import (
	"crypto/tls"
	"fmt"
	"github.com/mark-by/proxy/application"
	"github.com/mark-by/proxy/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
)

func Server(app *application.App) {
	server := http.Server{
		Addr:         fmt.Sprintf("%s:%s", viper.GetString(config.ProxyIP), viper.GetString(config.ProxyPort)),
		Handler:      http.HandlerFunc(app.Requests.Intercept),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	err := server.ListenAndServe()

	if err != nil {
		logrus.Fatalf("Сервер не запустился с ошибкой: %s", err)
	}
}
