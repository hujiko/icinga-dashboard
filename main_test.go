package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/hujiko/icinga-dashboard/icinga2apiclient"
)

type stubDashboardClient struct {
	appStatus *icinga2apiclient.IcingaApplication
	appErr    error
	cibStatus *icinga2apiclient.CIBStatus
	cibErr    error
	services  []icinga2apiclient.Service
	hosts     []icinga2apiclient.Host
	hostsErr  error
}

func (s stubDashboardClient) GetIcingaApplicationStatus() (*icinga2apiclient.IcingaApplication, error) {
	return s.appStatus, s.appErr
}

func (s stubDashboardClient) GetCIBStatus() (*icinga2apiclient.CIBStatus, error) {
	return s.cibStatus, s.cibErr
}

func (s stubDashboardClient) GetServices(minState int, maxState int, minStateType int) ([]icinga2apiclient.Service, error) {
	return s.services, nil
}

func (s stubDashboardClient) GetHosts(minStateType int) ([]icinga2apiclient.Host, error) {
	return s.hosts, s.hostsErr
}

func TestBuildServiceListRecords(t *testing.T) {
	services := []icinga2apiclient.Service{
		{HostName: "host-b", ServiceName: "disk", State: 2, StateType: 1},
		{HostName: "host-a", ServiceName: "disk", State: 2, StateType: 1},
		{HostName: "host-c", ServiceName: "ping", State: 1, StateType: 0},
	}

	records := buildServiceListRecords(services)
	sort.Sort(ByState(records))

	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	aggregated := records[0]
	if aggregated.Name != "disk" {
		t.Fatalf("expected first record to be disk, got %q", aggregated.Name)
	}
	if aggregated.HostField != "2 Hosts" {
		t.Errorf("expected aggregated host field %q, got %q", "2 Hosts", aggregated.HostField)
	}
	if !aggregated.IsAggregated {
		t.Errorf("expected aggregated record to be marked aggregated")
	}
	if aggregated.AggregatedHostsCount != 2 {
		t.Errorf("expected aggregated host count 2, got %d", aggregated.AggregatedHostsCount)
	}
	expectedHosts := []string{"host-b", "host-a"}
	for i, host := range expectedHosts {
		if aggregated.AggregatedHosts[i] != host {
			t.Errorf("expected aggregated host %d to be %q, got %q", i, host, aggregated.AggregatedHosts[i])
		}
	}

	single := records[1]
	if single.Name != "ping" {
		t.Fatalf("expected second record to be ping, got %q", single.Name)
	}
	if single.HostField != "host-c" {
		t.Errorf("expected single host field %q, got %q", "host-c", single.HostField)
	}
	if single.IsAggregated {
		t.Errorf("expected single record not to be aggregated")
	}
	if single.AggregatedHostsCount != 1 {
		t.Errorf("expected single host count 1, got %d", single.AggregatedHostsCount)
	}
	if len(single.AggregatedHosts) != 1 || single.AggregatedHosts[0] != "host-c" {
		t.Errorf("expected single aggregated hosts to contain host-c, got %v", single.AggregatedHosts)
	}
}

func TestBuildPageVariables(t *testing.T) {
	originalClient := client
	originalMinState := defaultMinState
	originalMaxState := defaultMaxState
	originalMinStateType := defaultMinStateType
	originalBaseURL := baseURL
	originalNow := now
	defer func() {
		client = originalClient
		defaultMinState = originalMinState
		defaultMaxState = originalMaxState
		defaultMinStateType = originalMinStateType
		baseURL = originalBaseURL
		now = originalNow
	}()

	client = stubDashboardClient{
		appStatus: &icinga2apiclient.IcingaApplication{EnableNotifications: false},
		cibStatus: &icinga2apiclient.CIBStatus{
			NumHostsUp:          4,
			NumHostsDown:        1,
			NumServicesOk:       9,
			NumServicesWarning:  2,
			NumServicesCritical: 1,
			NumServicesUnknown:  0,
		},
		services: []icinga2apiclient.Service{
			{HostName: "host-b", ServiceName: "disk", State: 2, StateType: 1},
			{HostName: "host-a", ServiceName: "disk", State: 2, StateType: 1},
			{HostName: "host-z", ServiceName: "ping", State: 1, StateType: 0},
		},
		hosts: []icinga2apiclient.Host{
			{Name: "beta", State: 1, StateType: 1},
			{Name: "alpha", State: 2, StateType: 0},
		},
	}
	defaultMinState = 1
	defaultMaxState = 2
	defaultMinStateType = 0
	baseURL = "https://icinga.example.test"
	now = func() time.Time {
		return time.Date(2026, time.March, 11, 8, 9, 10, 0, time.UTC)
	}

	req := httptest.NewRequest(http.MethodGet, "/?minStateType=1&minState=2&maxState=3", nil)
	page := buildPageVariables(req)

	if page.Error != nil {
		t.Fatalf("expected no error, got %v", page.Error)
	}
	if page.TimeString != "08:09:10" {
		t.Errorf("expected time string 08:09:10, got %q", page.TimeString)
	}
	if page.Time.Unix() != 1773216550 {
		t.Errorf("expected unix timestamp 1773216550, got %d", page.Time.Unix())
	}
	if page.MinStateType != "Hard" || page.MinState != "Critical" || page.MaxState != "Unknown" {
		t.Errorf("unexpected state labels: %+v", page)
	}
	if page.BaseURL != "https://icinga.example.test" {
		t.Errorf("unexpected base URL %q", page.BaseURL)
	}
	if !page.NotificationsDisabled {
		t.Errorf("expected notifications to be marked disabled")
	}
	if len(page.ServiceRecords) != 2 {
		t.Fatalf("expected 2 service records, got %d", len(page.ServiceRecords))
	}
	if page.ServiceRecords[0].Name != "disk" || !page.ServiceRecords[0].IsAggregated {
		t.Errorf("unexpected first service record: %+v", page.ServiceRecords[0])
	}
	if len(page.HostRecords) != 2 || page.HostRecords[0].Name != "alpha" || page.HostRecords[1].Name != "beta" {
		t.Errorf("hosts not sorted as expected: %+v", page.HostRecords)
	}
}

func TestRenderJSON(t *testing.T) {
	originalClient := client
	originalMinState := defaultMinState
	originalMaxState := defaultMaxState
	originalMinStateType := defaultMinStateType
	originalBaseURL := baseURL
	originalNow := now
	defer func() {
		client = originalClient
		defaultMinState = originalMinState
		defaultMaxState = originalMaxState
		defaultMinStateType = originalMinStateType
		baseURL = originalBaseURL
		now = originalNow
	}()

	client = stubDashboardClient{
		appStatus: &icinga2apiclient.IcingaApplication{EnableNotifications: true},
		cibStatus: &icinga2apiclient.CIBStatus{},
	}
	defaultMinState = 1
	defaultMaxState = 2
	defaultMinStateType = 0
	baseURL = "https://icinga.example.test"
	now = func() time.Time {
		return time.Date(2026, time.March, 11, 8, 9, 10, 0, time.UTC)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	rec := httptest.NewRecorder()

	renderJSON(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if contentType := rec.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected JSON content type, got %q", contentType)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "\"base_url\":\"https://icinga.example.test\"") {
		t.Errorf("expected base_url in JSON response, got %s", body)
	}
	if !strings.Contains(body, "\"notifications_disabled\":false") {
		t.Errorf("expected notifications_disabled=false in JSON response, got %s", body)
	}
}

func TestRenderDashboardReturnsInternalServerErrorOnDataError(t *testing.T) {
	originalClient := client
	originalMinState := defaultMinState
	originalMaxState := defaultMaxState
	originalMinStateType := defaultMinStateType
	originalBaseURL := baseURL
	originalNow := now
	defer func() {
		client = originalClient
		defaultMinState = originalMinState
		defaultMaxState = originalMaxState
		defaultMinStateType = originalMinStateType
		baseURL = originalBaseURL
		now = originalNow
	}()

	client = stubDashboardClient{
		appErr: errors.New("icinga unavailable"),
	}
	defaultMinState = 1
	defaultMaxState = 2
	defaultMinStateType = 0
	baseURL = "https://icinga.example.test"
	now = func() time.Time {
		return time.Date(2026, time.March, 11, 8, 9, 10, 0, time.UTC)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	renderDashboard(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "icinga unavailable") {
		t.Errorf("expected error text in rendered dashboard, got %s", rec.Body.String())
	}
}

func TestParseEnvVariables(t *testing.T) {
	t.Setenv("LISTEN_ADDRESS", ":9090")
	t.Setenv("ICINGA2_BASE_URL", "https://icinga.example.test")
	t.Setenv("ICINGA2_API_URL", "https://icinga-api.example.test")
	t.Setenv("ICINGA2_API_TIMEOUT", "10")
	t.Setenv("MIN_STATE", "2")
	t.Setenv("ICINGA2_API_VALIDATE_CERTIFICATE", "invalid")

	env := parseEnvVariables()

	if env["LISTEN_ADDRESS"] != ":9090" {
		t.Errorf("unexpected LISTEN_ADDRESS: %v", env["LISTEN_ADDRESS"])
	}
	if env["ICINGA2_API_TIMEOUT"] != 10 {
		t.Errorf("unexpected timeout: %v", env["ICINGA2_API_TIMEOUT"])
	}
	if env["MIN_STATE"] != 2 {
		t.Errorf("unexpected MIN_STATE: %v", env["MIN_STATE"])
	}
	if env["ICINGA2_API_VALIDATE_CERTIFICATE"] != 1 {
		t.Errorf("expected invalid int to fall back to default, got %v", env["ICINGA2_API_VALIDATE_CERTIFICATE"])
	}
	if env["MAX_STATE"] != 2 {
		t.Errorf("expected MAX_STATE default 2, got %v", env["MAX_STATE"])
	}
}

func TestStateNumToString(t *testing.T) {
	if stateNumToString(0) != "OK" {
		t.Errorf("expected OK for state 0")
	}
	if stateNumToString(3) != "Unknown" {
		t.Errorf("expected Unknown for state 3")
	}
	if stateNumToString(4) != "---" {
		t.Errorf("expected fallback for invalid state")
	}
}

func TestStateTypeNumToString(t *testing.T) {
	if stateTypeNumToString(0) != "Soft" {
		t.Errorf("expected Soft for state type 0")
	}
	if stateTypeNumToString(1) != "Hard" {
		t.Errorf("expected Hard for state type 1")
	}
	if stateTypeNumToString(-1) != "---" {
		t.Errorf("expected fallback for invalid state type")
	}
}
