package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"scf-proxy/pkg/mitm"
	"scf-proxy/pkg/scf"
	"sync"
	"time"
)

const (
	HTTP_ADDR         = "127.0.0.1:8080"
	SESSIONS_TO_CACHE = 10000
)

var (
	exampleWg  sync.WaitGroup
	clientPort string
)

func init() {
	flag.StringVar(&clientPort, "p", "8080", "scf-proxy 客户端端口")
	flag.StringVar(&scf.ScfApiProxyUrl, "scfurl", "", "scf-proxy 服务端地址")
	flag.Parse()
	if scf.ScfApiProxyUrl == "" {
		panic("scf-proxy 服务端地址为空")
	}
	fmt.Println(scf.ScfApiProxyUrl)

}

func main() {
	exampleWg.Add(1)
	runHTTPServer()
	// Uncomment the below line to keep the server running
	exampleWg.Wait()

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
			log.Printf("Processing request to: %s", req.URL)
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				// Use a TLS session cache to minimize TLS connection establishment
				// Requires Go 1.3+
				ClientSessionCache: tls.NewLRUClientSessionCache(SESSIONS_TO_CACHE),
			},
		},
	}

	handler, err := mitm.Wrap(rp, cryptoConfig)
	if err != nil {
		log.Fatalf("Unable to wrap reverse proxy: %s", err)
	}

	server := &http.Server{
		Addr:         ":" + clientPort,
		Handler:      handler,
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
	}

	go func() {
		log.Printf("About to start HTTP proxy at :%s", clientPort)
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Unable to start HTTP proxy: %s", err)
		}
		exampleWg.Done()
	}()

	return
}
