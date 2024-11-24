package icinga2apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

	url := fmt.Sprintf("%s/v1/objects/hosts", client.Hostname)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	// Icinga wants a GET, but GET requests can't contain a payload
	req.Header.Set("X-HTTP-Method-Override", "GET")

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

	var responseStruct getHostsResponse
	json.Unmarshal(body, &responseStruct)

	var hosts []Host
	for _, hostJson := range responseStruct.Results {
		hosts = append(hosts, NewHostFromJson(hostJson))
	}

	return hosts, nil
}

func NewHostFromJson(hostJson icinga2HostJson) Host {
	host := Host{
		Name:      hostJson.Name,
		State:     hostJson.Attributes.State,
		StateType: hostJson.Attributes.StateType,
	}

	return host
}
