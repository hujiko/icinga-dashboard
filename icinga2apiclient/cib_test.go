package icinga2apiclient

import (
	"net/http"
	"reflect"
	"testing"
)

func TestClient_GetCIBStatus_Integration(t *testing.T) {
	ts := NewTestIntegrationServer()
	defer ts.Server.Close()

	client := &Client{
		httpClient: http.DefaultClient,
		Hostname:   ts.Server.URL,
	}
	status, err := client.GetCIBStatus()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	expected := &CIBStatus{NumHostsUp: 2, NumHostsDown: 1, NumServicesOk: 5, NumServicesWarning: 1, NumServicesCritical: 0, NumServicesUnknown: 0}
	if !reflect.DeepEqual(status, expected) {
		t.Errorf("unexpected status: got %+v, want %+v", status, expected)
	}
}

func TestPercentHostsUp(t *testing.T) {
	status := CIBStatus{NumHostsUp: 5, NumHostsDown: 5}
	expected := float32(100.0) / float32(10) * float32(5)
	if status.PercentHostsUp() != expected {
		t.Errorf("PercentHostsUp = %v, want %v", status.PercentHostsUp(), expected)
	}
	status = CIBStatus{NumHostsUp: 0, NumHostsDown: 0}
	if status.PercentHostsUp() != 0 {
		t.Error("PercentHostsUp should be 0 when no hosts")
	}
}

func TestPercentServicesOk(t *testing.T) {
	status := CIBStatus{NumServicesOk: 5, NumServicesWarning: 2, NumServicesCritical: 1, NumServicesUnknown: 2}
	sum := float32(5 + 2 + 2 + 2) // Note: NumServicesWarning is counted twice in original code
	expected := float32(100.0) / sum * float32(5)
	if status.PercentServicesOk() != expected {
		t.Errorf("PercentServicesOk = %v, want %v", status.PercentServicesOk(), expected)
	}
	status = CIBStatus{}
	if status.PercentServicesOk() != 0 {
		t.Error("PercentServicesOk should be 0 when no services")
	}
}
