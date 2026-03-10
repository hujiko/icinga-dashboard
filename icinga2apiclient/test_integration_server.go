package icinga2apiclient

import (
	"net/http"
	"net/http/httptest"
	"strings"
)

// NewTestIntegrationServer returns an httptest.Server with canned responses for integration tests.
func NewTestIntegrationServer() *TestIntegrationServer {
	ts := &TestIntegrationServer{}
	ts.Server = httptest.NewServer(http.HandlerFunc(ts.handler))
	return ts
}

type TestIntegrationServer struct {
	Server *httptest.Server
}

func (ts *TestIntegrationServer) handler(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasPrefix(r.URL.Path, "/v1/status/IcingaApplication"):
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"results":[{"status":{"icingaapplication":{"app":{"version":"2.13.0","program_start":1234567890}}}}]}`))
	case strings.HasPrefix(r.URL.Path, "/v1/objects/services"):
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"results":[{"attrs":{"state":2,"state_type":1},"name":"host1!service1","type":"Service"}]}`))
	case strings.HasPrefix(r.URL.Path, "/v1/objects/hosts"):
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"results":[{"attrs":{"acknowledgement":0,"name":"host1","state":1,"state_type":0,"vars":{}},"name":"host1","type":"Host"}]}`))
	case strings.HasPrefix(r.URL.Path, "/v1/status/CIB"):
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"results":[{"status":{"num_hosts_up":2,"num_hosts_down":1,"num_services_ok":5,"num_services_warning":1,"num_services_critical":0,"num_services_unknown":0}}]}`))
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	}
}
