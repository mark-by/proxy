package entity

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
)

type Request struct {
	ID  int `json:"id"`
	Raw string `json:"raw"`
	URL string `json:"url"`
}

func (r Request) Revive() (*http.Request, error) {
	buff := bytes.Buffer{}
	buff.Write([]byte(r.Raw))
	request, err := http.ReadRequest(bufio.NewReader(&buff))
	if err != nil {
		return nil, err
	}

	request.Header.Del("Proxy-Connection")
	if len(r.URL) != 0 && r.URL[0] == '/' {
		r.URL = fmt.Sprintf("https://%s%s", request.Host, r.URL)
	}

	newRequest, err := http.NewRequest(request.Method, r.URL, request.Body)
	copyHeaders(request, newRequest)
	return newRequest, nil
}

func copyHeaders(src *http.Request, dst *http.Request) {
	for header, values := range src.Header {
		for _, value := range values {
			dst.Header.Add(header, value)
		}
	}
}
