package main

import (
	"crypto/tls"
	"github.com/chroblert/jlog"
	"golang.org/x/net/http2"
	"net"
	"net/http"
	"net/http/httputil"
	"scf-proxy/pkg/mitm"
	"scf-proxy/pkg/scf"
	"scf-proxy/pkg/viper"
	"sync"
	"time"
)

const (
	//HTTP_ADDR         = "127.0.0.1:8080"
	SESSIONS_TO_CACHE = 10000
)

var (
	exampleWg  sync.WaitGroup
	clientPort string
)

func init() {
	_ = jlog.SetLogFullPath("logs/client.log", 0755, 0755)
	jlog.SetStoreToFile(true)
	jlog.SetMaxSizePerLogFile("10MB")
	jlog.SetMaxStoreDays(30)
	jlog.SetLogCount(30)
	jlog.IsIniCreateNewLog(true)
	jlog.SetVerbose(true)
	scf.ScfApiProxyUrl, clientPort = viper.YamlConfig()
}

func main() {
	exampleWg.Add(1)
	runHTTPServer()
	// Uncomment the below line to keep the server running
	exampleWg.Wait()
	//defer jlog.Flush()

	// Output:

}

func runHTTPServer() {
	cryptoConfig := &mitm.CryptoConfig{
		PKFile:   "proxypk.pem",
		CertFile: "proxycert.pem",
		ServerTLSConfig: &tls.Config{
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
				tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
				tls.TLS_RSA_WITH_RC4_128_SHA,
				tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
				tls.TLS_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			},
			PreferServerCipherSuites: true,
		},
	}

	rp := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			jlog.Infof("Processing request to: %s", req.URL)
		},
		//Transport: &http.Transport{
		//	//AllowHTTP: true,
		//	TLSClientConfig: &tls.Config{
		//		// Use a TLS session cache to minimize TLS connection establishment
		//		// Requires Go 1.3+
		//		ClientSessionCache: tls.NewLRUClientSessionCache(SESSIONS_TO_CACHE),
		//	},
		//	//DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
		//	//	return net.Dial(network, addr)
		//	//},
		//},
		Transport: &http2.Transport{
			AllowHTTP: true,
			TLSClientConfig: &tls.Config{
				// Use a TLS session cache to minimize TLS connection establishment
				// Requires Go 1.3+
				ClientSessionCache: tls.NewLRUClientSessionCache(SESSIONS_TO_CACHE),
			},
			DialTLS: func(netw, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(netw, addr)
			},
		},
		FlushInterval: -1,
	}

	handler, err := mitm.Wrap(rp, cryptoConfig)
	if err != nil {
		jlog.Fatalf("Unable to wrap reverse proxy: %s", err)
	}

	server := &http.Server{
		Addr:         ":" + clientPort,
		Handler:      handler,
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
	}
	_ = http2.ConfigureServer(server, nil)
	go func() {
		jlog.Infof("About to start HTTP proxy at :%s\n", clientPort)
		if err := server.ListenAndServe(); err != nil {
			jlog.Fatalf("Unable to start HTTP proxy: %s", err)
		}
		exampleWg.Done()
	}()

	return
	//}
}
