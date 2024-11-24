package icinga2apiclient

import (
	"net/http"
)

type Client struct {
	httpClient *http.Client
	Hostname   string
	Username   string
	Password   string
}

// "attrs":  []string{"name", "state", "state_type", "downtime_depth", "acknowledgement", "vars"},
// "joins":  []string{},
// "filter": []string{"host.state != 0", "host.downtime_depth == 0", "host.acknowledgement == 0", "'host.state_type >= 1"},
type requestPayload struct {
	Attributes []string `json:"attrs"`
	Joins      []string `json:"joins"`
	Filters    string   `json:"filter"`
}

type getServiceResponse struct {
	Results []icinga2ServiceJson `json:"results"`
}

type getHostsResponse struct {
	Results []icinga2HostJson `json:"results"`
}

// {"attrs":{"acknowledgement":0,"display_name":"disk-timeouts","downtime_depth":0,"name":"disk-timeouts","state":3,
// "state_type":1,"vars":{"oncall":"plaser"}},"joins":{},"meta":{},"name":"keepalived-1.graylog-coresec.ams1!disk-timeouts","type":"Service"}
type icinga2ServiceAttributesJson struct {
	Acknowledgement int                    `json:"acknowledgement"`
	DisplayName     string                 `json:"display_name"`
	DowntimeDepth   int                    `json:"downtime_depth"`
	Name            string                 `json:"name"`
	State           int                    `json:"state"`
	StateType       int                    `json:"state_type"`
	Vars            map[string]interface{} `json:"vars"`
}

type icinga2HostAttributesJson struct {
	Acknowledgement int                    `json:"acknowledgement"`
	Name            string                 `json:"name"`
	State           int                    `json:"state"`
	StateType       int                    `json:"state_type"`
	Vars            map[string]interface{} `json:"vars"`
}

type icinga2ServiceJson struct {
	Attributes icinga2ServiceAttributesJson `json:"attrs"`
	Name       string                       `json:"name"`
	Type       string                       `json:"type"`
}

type icinga2HostJson struct {
	Attributes icinga2HostAttributesJson `json:"attrs"`
	Name       string                    `json:"name"`
	Type       string                    `json:"type"`
}

type Service struct {
	HostName    string
	ServiceName string
	State       int
	StateType   int
}

type Host struct {
	Name      string
	State     int
	StateType int
}