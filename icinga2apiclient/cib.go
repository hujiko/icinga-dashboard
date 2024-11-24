package icinga2apiclient

import (
	"encoding/json"
	"fmt"
	"net/http"
)

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
	responseBody, err := client.makeRequest(http.MethodGet, "/v1/status/CIB", nil)
	if err != nil {
		fmt.Printf("Error fetching CIB status: %v\n", err)
		return nil, err
	}

	var cibStatusResponseStruct cibStatusResponse
	err = json.Unmarshal(responseBody, &cibStatusResponseStruct)
	if err != nil {
		fmt.Printf("Unable to parse JSON: %v\n", err)
		return nil, err
	}

	return &cibStatusResponseStruct.Results[0].Status, nil
}
