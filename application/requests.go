package application

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/mark-by/proxy/domain/repository"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"sync"
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

	defer func() {
		if err = resp.Body.Close(); err != nil {
			logrus.Error("fail to close body:", err)
		}
	}()

	copyResponse(resp, w)
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

func (requests Requests) tunnel(w http.ResponseWriter, r *http.Request) {
	name, _, err := net.SplitHostPort(r.Host)
	if err != nil || name == "" {
		logrus.Error("cannot determine certificate name for ", r.Host)
		http.Error(w, "no upstream", http.StatusServiceUnavailable)
		return
	}

	certificate, err := getCertificateByName(name)
	if err != nil {
		logrus.Error("fail to get certificate: ", err)
		http.Error(w, "nu upstream", http.StatusNotImplemented)
		return
	}

	var sTlsConn *tls.Conn
	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS10,
		Certificates: []tls.Certificate{*certificate},
		GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			cLocalTlsConfig := &tls.Config{
				InsecureSkipVerify: true,
				ServerName:         hello.ServerName,
			}
			sTlsConn, err = tls.Dial("tcp", r.Host, cLocalTlsConfig)
			if err != nil {
				log.Println("dial", r.Host, err)
				return nil, err
			}
			return getCertificateByName(hello.ServerName)
		},
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	cRawConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	_, err = cRawConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n" +
		"Proxy-agent: Golang-Proxy\r\n" +
		"\r\n"))
	if err != nil {
		logrus.Error("cRawConn fail to handshake for ", r.Host,": ", err)
		cRawConn.Close()
		return
	}

	cTlsConn := tls.Server(cRawConn, tlsConfig)
	err = cTlsConn.Handshake()
	if err != nil {
		logrus.Error("fail to handshake for ", r.Host,": ", err)
		cTlsConn.Close()
		cRawConn.Close()
		return
	}
	defer cTlsConn.Close()

	if sTlsConn == nil {
		logrus.Error("fail to determine cert name for ", r.Host)
		return
	}
	defer sTlsConn.Close()

	rp := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Host = r.Host
			r.URL.Scheme = "https"
		},
		Transport: &http.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				if sTlsConn == nil {
					return nil, errors.New("closed on dial")
				}
				return sTlsConn, nil
			},
		},
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	wc := &onCloseConn{cTlsConn, func() { wg.Done() }}
	http.Serve(&oneShotListener{wc}, wrap(rp))
	wg.Wait()
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
