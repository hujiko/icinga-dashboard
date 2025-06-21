package icinga2apiclient

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
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

func (client *Client) makeRequest(verb string, path string, payload []byte) ([]byte, error) {
	url := client.Hostname + path
	req, err := http.NewRequest(verb, url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	// Icinga wants a GET, but GET requests can't contain a payload
	req.Header.Set("X-HTTP-Method-Override", "GET")

	if client.Username != "" && client.Password != "" {
		req.SetBasicAuth(client.Username, client.Password)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 299 {
		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       string(body),
		}
	}

	return body, nil
}
