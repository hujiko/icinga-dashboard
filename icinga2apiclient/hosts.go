package icinga2apiclient

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (client *Client) GetHosts(minStateType int) ([]Host, error) {
	payload := requestPayload{
		Attributes: []string{"name", "state", "state_type", "downtime_depth", "acknowledgement", "vars"},
		Filters:    fmt.Sprintf("host.state != 0 && host.downtime_depth == 0 && host.acknowledgement == 0 && host.state_type >= %d", minStateType),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling JSON payload: %v\n", err)
		return nil, err
	}

	responseBody, err := client.makeRequest(http.MethodPost, "/v1/objects/hosts", jsonPayload)
	if err != nil {
		fmt.Printf("Error fetching hosts: %v\n", err)
		return nil, err
	}

	var responseStruct getHostsResponse
	err = json.Unmarshal(responseBody, &responseStruct)
	if err != nil {
		fmt.Printf("Unable to parse JSON: %v\n", err)
		return nil, err
	}

	var hosts []Host
	for _, hostJSON := range responseStruct.Results {
		hosts = append(hosts, NewHostFromJSON(hostJSON))
	}

	return hosts, nil
}

func NewHostFromJSON(hostJSON icinga2hostJSON) Host {
	host := Host{
		Name:      hostJSON.Name,
		State:     hostJSON.Attributes.State,
		StateType: hostJSON.Attributes.StateType,
	}

	return host
}
