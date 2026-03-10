package icinga2apiclient

import (
	"net/http"
	"testing"
)

func TestClient_GetServices_Integration(t *testing.T) {
	ts := NewTestIntegrationServer()
	defer ts.Server.Close()

	client := &Client{
		httpClient: http.DefaultClient,
		Hostname:   ts.Server.URL,
	}
	services, err := client.GetServices(0, 3, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services))
	}
	if services[0].HostName != "host1" || services[0].ServiceName != "service1" || services[0].State != 2 || services[0].StateType != 1 {
		t.Errorf("unexpected service: %+v", services[0])
	}
}

func TestNewServiceFromJSON(t *testing.T) {
	json := icinga2serviceJSON{
		Attributes: icinga2ServiceAttributesJSON{
			State:     2,
			StateType: 1,
		},
		Name: "host1!service1",
		Type: "Service",
	}
	service := NewServiceFromJSON(json)
	if service.HostName != "host1" || service.ServiceName != "service1" {
		t.Errorf("NewServiceFromJSON failed: got HostName=%v, ServiceName=%v", service.HostName, service.ServiceName)
	}
	if service.State != 2 || service.StateType != 1 {
		t.Errorf("NewServiceFromJSON failed: got State=%v, StateType=%v", service.State, service.StateType)
	}
}
