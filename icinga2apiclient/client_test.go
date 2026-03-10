package icinga2apiclient

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient("https://example.com", "", "", "", 5, true)
	if err != nil {
		t.Errorf("NewClient returned error: %v", err)
	}
	if client == nil {
		t.Error("NewClient returned nil client")
	}
}

func TestHTTPError_Error(t *testing.T) {
	err := &HTTPError{StatusCode: 404, Status: "Not Found", Body: "missing"}
	expected := "HTTP 404: Not Found - missing"
	if err.Error() != expected {
		t.Errorf("HTTPError.Error() = %v, want %v", err.Error(), expected)
	}
}
