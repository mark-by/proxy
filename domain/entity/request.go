package entity

import (
	"bufio"
	"bytes"
	"net/http"
)

type Request struct {
	ID  uint64 `json:"id"`
	Raw string `json:"raw"`
}

func (r Request) Revive() (*http.Request, error) {
	buff := bytes.Buffer{}
	buff.Write([]byte(r.Raw))
	return http.ReadRequest(bufio.NewReader(&buff))
}
