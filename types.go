package main

import (
	"net/url"
	"strconv"
	"time"

	"github.com/hujiko/icinga-dashboard/icinga2apiclient"
)

type PageVariables struct {
	TimeString            string                      `json:"time_string"`
	Time                  timestamp                   `json:"timestamp"`
	ServiceRecords        []PageServiceListRecord     `json:"services"`
	HostRecords           []PageHostListRecord        `json:"hosts"`
	CIBStatus             *icinga2apiclient.CIBStatus `json:"cib_status"`
	MinStateType          string                      `json:"min_state_type"`
	MinState              string                      `json:"min_state"`
	MaxState              string                      `json:"max_state"`
	Error                 error                       `json:"error"`
	BaseURL               string                      `json:"base_url"`
	NotificationsDisabled bool                        `json:"notifications_disabled"`
}

type PageServiceListRecord struct {
	Name                 string   `json:"name"`
	HostField            string   `json:"host_field"`
	State                int      `json:"state"`
	StateType            int      `json:"state_type"`
	IsAggregated         bool     `json:"is_aggregated"`
	AggregatedHosts      []string `json:"aggregated_hosts"`
	AggregatedHostsCount int      `json:"aggregated_hosts_count"`
}

type PageHostListRecord struct {
	State     int    `json:"state"`
	StateType int    `json:"state_type"`
	Name      string `json:"name"`
}

func (r *PageHostListRecord) URLEncodedHost() string {
	return url.QueryEscape(r.Name)
}

func (r *PageServiceListRecord) URLEncodedHost() string {
	return url.QueryEscape(r.HostField)
}

func (r *PageServiceListRecord) URLEncodedService() string {
	return url.QueryEscape(r.Name)
}

// Allows sorting services by state
type ByState []PageServiceListRecord

func (a ByState) Len() int { return len(a) }
func (a ByState) Less(i, j int) bool {
	// Sort by State descending
	if a[i].State != a[j].State {
		return a[i].State > a[j].State // Descending order
	}
	// If States are equal, sort by ServiceName ascending
	if a[i].Name != a[j].Name {
		return a[i].Name < a[j].Name
	}

	return a[i].StateType < a[j].StateType // Ascending order
}
func (a ByState) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Allows sorting hosts by name
type ByName []PageHostListRecord

func (a ByName) Len() int { return len(a) }
func (a ByName) Less(i, j int) bool {
	return a[i].Name < a[j].Name // Ascending order
}
func (a ByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

type timestamp struct {
	time.Time
}

func (t timestamp) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(time.Time(t.Time).Unix(), 10)), nil
}

func (t *timestamp) UnmarshalJSON(data []byte) error {
	i, err := strconv.ParseInt(string(data[:]), 10, 64)
	if err != nil {
		return err
	}
	*t = timestamp{
		time.Unix(i, 0),
	}
	return nil
}
