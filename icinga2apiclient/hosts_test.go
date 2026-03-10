package icinga2apiclient

import (
	"net/http"
	"testing"
)

func TestClient_GetHosts_Integration(t *testing.T) {
	ts := NewTestIntegrationServer()
	defer ts.Server.Close()

	client := &Client{
		httpClient: http.DefaultClient,
		Hostname:   ts.Server.URL,
	}
	hosts, err := client.GetHosts(0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(hosts))
	}
	if hosts[0].Name != "host1" || hosts[0].State != 1 || hosts[0].StateType != 0 {
		t.Errorf("unexpected host: %+v", hosts[0])
	}
}

func TestNewHostFromJSON(t *testing.T) {
	json := icinga2hostJSON{
		Attributes: icinga2HostAttributesJSON{
			State:     1,
			StateType: 0,
		},
		Name: "host1",
		Type: "Host",
	}
	host := NewHostFromJSON(json)
	if host.Name != "host1" {
		t.Errorf("NewHostFromJSON failed: got Name=%v", host.Name)
	}
	if host.State != 1 || host.StateType != 0 {
		t.Errorf("NewHostFromJSON failed: got State=%v, StateType=%v", host.State, host.StateType)
	}
}
