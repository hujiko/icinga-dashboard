package icinga2apiclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"
)

func NewClient(hostName string, certFile string, keyFile string, caCertFile string, timeOutSecs int, verifyCertificate bool) (*Client, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: !verifyCertificate,
	}

	if certFile != "" && keyFile != "" {
		// Load client certificate
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			fmt.Printf("Error loading client certificate: %v\n", err)
			return nil, err
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if caCertFile != "" {
		// Load CA certificate
		caCert, err := os.ReadFile(caCertFile)
		if err != nil {
			fmt.Printf("Error reading CA certificate: %v\n", err)
			return nil, err
		}

		// Create a CA certificate pool and add the CA certificate
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tlsConfig.RootCAs = caCertPool
	}

	// Create an HTTP client with the TLS config
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: time.Duration(timeOutSecs) * time.Second,
	}

	return &Client{
		httpClient: client,
		Hostname:   hostName,
	}, nil
}
