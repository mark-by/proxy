package application

import (
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
)

func CopyResponse(src *http.Response, dst http.ResponseWriter) {
	for name, values := range src.Header {
		dst.Header()[name] = values
	}

	dst.WriteHeader(src.StatusCode)

	if _, err := io.Copy(dst, src.Body); err != nil {
		logrus.Error("fail to write body:", err)
	}
}