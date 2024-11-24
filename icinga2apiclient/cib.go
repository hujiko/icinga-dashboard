package icinga2apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type cibStatusResponse struct {
	Results []cIBStatusResult `json:"results"`
}
type cIBStatusResult struct {
	Name   string    `json:"name"`
	Status CIBStatus `json:"status"`
}

type CIBStatus struct {
	NumHostsUp          int `json:"num_hosts_up"`
	NumHostsDown        int `json:"num_hosts_down"`
	NumServicesOk       int `json:"num_services_ok"`
	NumServicesWarning  int `json:"num_services_warning"`
	NumServicesCritical int `json:"num_services_critical"`
	NumServicesUnknown  int `json:"num_services_unknown"`
}

func (s *CIBStatus) PercentHostsUp() float32 {
	sumOfHosts := float32(s.NumHostsUp + s.NumHostsDown)
	if sumOfHosts == 0 {
		return 0
	}
	return float32(100.0) / sumOfHosts * float32(s.NumHostsUp)
}

func (s *CIBStatus) PercentServicesOk() float32 {
	sumOfServices := float32(s.NumServicesOk + s.NumServicesWarning + s.NumServicesWarning + s.NumServicesUnknown)
	if sumOfServices == 0 {
		return 0
	}
	return float32(100.0) / sumOfServices * float32(s.NumServicesOk)
}

func (client *Client) GetCIBStatus() (*CIBStatus, error) {
	url := fmt.Sprintf("%s/v1/status/CIB", client.Hostname)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	if client.Username != "" && client.Password != "" {
		req.SetBasicAuth(client.Username, client.Password)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return nil, err
	}

	var cibStatusResponseStruct cibStatusResponse
	json.Unmarshal(body, &cibStatusResponseStruct)

	return &cibStatusResponseStruct.Results[0].Status, nil
}
