package icinga2apiclient

import (
	"net/http"
	"testing"
)

func TestClient_GetIcingaApplicationStatus_Integration(t *testing.T) {
	ts := NewTestIntegrationServer()
	defer ts.Server.Close()

	client := &Client{
		httpClient: http.DefaultClient,
		Hostname:   ts.Server.URL,
	}
	app, err := client.GetIcingaApplicationStatus()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if app == nil {
		t.Fatalf("expected application status, got nil")
	}
	if app.Version != "2.13.0" || app.ProgramStart != 1234567890 {
		t.Errorf("unexpected application status: %+v", app)
	}
}
