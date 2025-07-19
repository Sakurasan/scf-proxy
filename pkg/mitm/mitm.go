package mitm

import (
	"crypto/tls"
	"fmt"
	"github.com/chroblert/jlog"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"io"
	"net/http"
	"scf-proxy/pkg/scf"
	"strings"
	"sync"
	"time"
)

const (
	CONNECT = "CONNECT"
)

// HandlerWrapper wraps an http.Handler with MITM'ing functionality
type HandlerWrapper struct {
	cryptoConf      *CryptoConfig
	wrapped         http.Handler
	pk              *PrivateKey
	pkPem           []byte
	issuingCert     *Certificate
	issuingCertPem  []byte
	serverTLSConfig *tls.Config
	dynamicCerts    *Cache
	certMutex       sync.Mutex
}

func Wrap(handler http.Handler, cryptoConf *CryptoConfig) (*HandlerWrapper, error) {
	h2s := &http2.Server{
		IdleTimeout: time.Second * 60,
	}
	wrapper := &HandlerWrapper{
		cryptoConf: cryptoConf,
		//wrapped:      handler,
		//http/2 支持
		wrapped:      h2c.NewHandler(handler, h2s),
		dynamicCerts: NewCache(),
	}
	err := wrapper.initCrypto()
	if err != nil {
		return nil, err
	}
	return wrapper, nil
}

// ServeHTTP implements ServeHTTP from http.Handler
func (wrapper *HandlerWrapper) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if req.Method == CONNECT {
		wrapper.intercept(resp, req)
	} else {
		// wrapper.wrapped.ServeHTTP(resp, req)
		//reqdump, _ := httputil.DumpRequest(req, true)
		//fmt.Println("dump req:", string(reqdump))
		//jlog.Info(string(reqdump))
		scf.HandlerHttp(resp, req)
	}
}

func (wrapper *HandlerWrapper) intercept(resp http.ResponseWriter, req *http.Request) {
	// Find out which host to MITM
	addr := hostIncludingPort(req)
	host := strings.Split(addr, ":")[0]

	cert, err := wrapper.mitmCertForName(host)
	if err != nil {
		msg := fmt.Sprintf("Could not get mitm cert for name: %s\nerror: %s", host, err)
		respBadGateway(resp, msg)
		return
	}

	connIn, _, err := resp.(http.Hijacker).Hijack()
	if err != nil {
		msg := fmt.Sprintf("Unable to access underlying connection from client: %s", err)
		respBadGateway(resp, msg)
		return
	}
	tlsConfig := makeConfig(wrapper.cryptoConf.ServerTLSConfig)
	//HTTP/2 支持
	tlsConfig.NextProtos = []string{"h2"}

	tlsConfig.Certificates = []tls.Certificate{*cert}
	tlsConnIn := tls.Server(connIn, tlsConfig)

	listener := &mitmListener{tlsConnIn}

	handler := http.HandlerFunc(func(resp2 http.ResponseWriter, req2 *http.Request) {
		//req2.ProtoMajor
		req2.URL.Scheme = "https"
		req2.URL.Host = req2.Host
		req2.RequestURI = req2.URL.String()
		// wrapper.wrapped.ServeHTTP(resp2, req2)

		scf.HandlerHttp(resp2, req2)
	})

	go func() {
		err = http.Serve(listener, handler)
		if err != nil && err != io.EOF {
			jlog.Printf("Error serving mitm'ed connection: %s", err)
		}
	}()

	connIn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
}

func makeConfig(template *tls.Config) *tls.Config {
	tlsConfig := &tls.Config{}
	if template != nil {
		// Copy the provided tlsConfig
		*tlsConfig = *template
	}
	return tlsConfig
}

func hostIncludingPort(req *http.Request) (host string) {
	host = req.Host
	if !strings.Contains(host, ":") {
		host = host + ":443"
	}
	return
}

func respBadGateway(resp http.ResponseWriter, msg string) {
	jlog.Error(msg)
	resp.WriteHeader(502)
	resp.Write([]byte(msg))
}
