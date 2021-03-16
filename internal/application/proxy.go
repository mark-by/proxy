package application

import (
	"bufio"
	"crypto/tls"
	"errors"
	"github.com/mark-by/proxy/internal/domain/entity"
	"github.com/mark-by/proxy/internal/domain/repository"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/http/httputil"
)

type Requests struct {
	repositories *repository.Repositories
}

func (requests Requests) Delete(id int) error {
	return requests.repositories.Requests.Delete(id)
}

func (requests Requests) DeleteAll() error {
	return requests.repositories.Requests.DeleteAll()
}

func (requests Requests) Get(id int) *entity.Request {
	req, err := requests.repositories.Requests.Get(id)
	if err != nil {
		logrus.Error("Get: ", err)
		return nil
	}
	return req
}

func (requests Requests) GetAll() []entity.Request {
	reqs, err := requests.repositories.Requests.List()
	if err != nil {
		logrus.Error("Get all : ", err)
		return nil
	}
	return reqs
}

func newRequests(repositories *repository.Repositories) *Requests {
	return &Requests{repositories}
}

func (requests Requests) Intercept(w http.ResponseWriter, r *http.Request) {
	logrus.Info("URI: ", r.RequestURI)
	if r.Method == http.MethodConnect {
		requests.tunnel(w, r)
		return
	}

	requests.proxy(w, r)
}

func (requests Requests) proxy(w http.ResponseWriter, r *http.Request) {
	go requests.saveRequest(r)
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	CopyResponse(resp, w)

	if err = resp.Body.Close(); err != nil {
		logrus.Error("fail to close body:", err)
	}
}

func (requests Requests) saveRequest(r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		logrus.Error("fail to dump request:", err)
	}

	logrus.Info("URI IN SAVE: ", r.RequestURI)

	_, err = requests.repositories.Requests.Save(r.RequestURI, string(dump))
	if err != nil {
		logrus.Error("fail to save request:", err)
	}
}

func (requests Requests) hijackConnect(w http.ResponseWriter) (net.Conn, error) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return nil, errors.New("hijacking not supported")
	}

	clientRawConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return nil, err
	}

	_, err = clientRawConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	if err != nil {
		logrus.Error("cRawConn fail to handshake ", err)
		clientRawConn.Close()
		return nil, err
	}
	return clientRawConn, nil
}

func (requests Requests) initializeTCPClientConn(conn net.Conn, w http.ResponseWriter, r *http.Request) (*tls.Conn, *tls.Config, error) {
	name, _, err := net.SplitHostPort(r.Host)
	if err != nil || name == "" {
		logrus.Error("cannot determine certificate name for ", r.Host)
		http.Error(w, "no upstream", http.StatusServiceUnavailable)
		return nil, nil, err
	}

	certificate, err := getCertificateByName(name)
	if err != nil {
		logrus.Error("fail to get certificate: ", err)
		http.Error(w, "nu upstream", http.StatusNotImplemented)
		return nil, nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*certificate},
		ServerName: r.URL.Scheme,
	}

	tcpConn := tls.Server(conn, tlsConfig)
	err = tcpConn.Handshake()
	if err != nil {
		tcpConn.Close()
		conn.Close()
		logrus.Error("fail to handshake: ", err)
		return nil, nil, err
	}

	return tcpConn, tlsConfig, nil
}

func (requests Requests) readRequestFromTCP(conn *tls.Conn) (*http.Request, error) {
	clientReader := bufio.NewReader(conn)
	request, err := http.ReadRequest(clientReader)
	if err != nil {
		logrus.Error("fail to read request: ", err)
		return nil, err
	}
	return request, nil
}

func (requests Requests) serveRequestByTCP(client *tls.Conn, server *tls.Conn, request *http.Request) error {
	dumpRequest, err := httputil.DumpRequest(request, true)
	if err != nil {
		logrus.Error("fail to dump request: ", err)
		return err
	}
	logrus.Info("HTTPS REQUEST: \n", string(dumpRequest))
	_, err = server.Write(dumpRequest)
	if err != nil {
		logrus.Error("fail to write request: ", err)
		return err
	}

	serverReader := bufio.NewReader(server)
	response, err := http.ReadResponse(serverReader, request)
	if err != nil {
		logrus.Error("fail to read response: ", err)
		return err
	}

	rawResponse, err := httputil.DumpResponse(response, true)
	if err != nil {
		logrus.Error("fail to dump response: ", err)
		return err
	}

	_, err = client.Write(rawResponse)
	if err != nil {
		logrus.Error("fail to write response: ", err)
		return err
	}

	return nil
}

func (requests Requests) tunnel(w http.ResponseWriter, r *http.Request) {
	clientRawConn, err := requests.hijackConnect(w)
	if err != nil {
		return
	}
	defer clientRawConn.Close()

	tcpClientConn, tlsConfig, err := requests.initializeTCPClientConn(clientRawConn, w, r)
	if err != nil {
		return
	}
	defer tcpClientConn.Close()

	request, err := requests.readRequestFromTCP(tcpClientConn)
	if err != nil {
		return
	}
	go requests.saveRequest(request)

	tcpServerConn, err := tls.Dial("tcp", r.URL.Host, tlsConfig)
	if err != nil {
		logrus.Error("fail to create tcp server conn: ", err)
		return
	}
	defer tcpServerConn.Close()

	err = requests.serveRequestByTCP(tcpClientConn, tcpServerConn, request)
	if err != nil {
		return
	}
}

var _ IRequests = &Requests{}
