package application

import (
	"bufio"
	"crypto/tls"
	"errors"
	"github.com/mark-by/proxy/domain/repository"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
)

type Requests struct {
	repositories *repository.Repositories
}

func newRequests(repositories *repository.Repositories) *Requests {
	return &Requests{repositories}
}

func (requests Requests) Intercept(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		requests.tunnel(w, r)
		return
	}

	requests.proxy(w, r)
}

func (requests Requests) proxy(w http.ResponseWriter, r *http.Request) {
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	copyResponse(resp, w)

	if err = resp.Body.Close(); err != nil {
		logrus.Error("fail to close body:", err)
	}
}

func (requests Requests) saveRequest(r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		logrus.Error("fail to dump request:", err)
	}

	_, err = requests.repositories.Requests.Save(string(dump))
	if err != nil {
		logrus.Error("fail to save request:", err)
	}
}

func copyResponse(src *http.Response, dst http.ResponseWriter) {
	for name, values := range src.Header {
		dst.Header()[name] = values
	}

	dst.WriteHeader(src.StatusCode)

	if _, err := io.Copy(dst, src.Body); err != nil {
		logrus.Error("fail to write body:", err)
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
	logrus.Info("HTTPS RESPONSE: \n", string(rawResponse))

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

func wrap(upstream http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dump, _ := httputil.DumpRequest(r, true)
		logrus.Info("IN WRAPPER: ", string(dump))

		upstream.ServeHTTP(w, r)
	})
}

type onCloseConn struct {
	net.Conn
	f func()
}

func (c *onCloseConn) Close() error {
	if c.f != nil {
		c.f()
		c.f = nil
	}
	return c.Conn.Close()
}

type oneShotListener struct {
	c net.Conn
}

func (l *oneShotListener) Accept() (net.Conn, error) {
	if l.c == nil {
		return nil, errors.New("closed on accept")
	}
	c := l.c
	l.c = nil
	return c, nil
}

func (l *oneShotListener) Close() error {
	return nil
}

func (l *oneShotListener) Addr() net.Addr {
	return l.c.LocalAddr()
}

var _ IRequests = &Requests{}
