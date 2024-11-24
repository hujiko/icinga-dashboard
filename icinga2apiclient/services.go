package icinga2apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (client *Client) GetServices(minState int, maxState int, minStateType int) ([]Service, error) {
	payload := requestPayload{
		Attributes: []string{"name", "state", "state_type", "downtime_depth", "acknowledgement", "vars", "display_name"},
		Filters:    fmt.Sprintf("service.state >= %d && service.state <= %d && service.state_type >= %d && service.acknowledgement == 0 && service.downtime_depth == 0 && host.state == 0", minState, maxState, minStateType),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling JSON payload: %v\n", err)
		return nil, err
	}

	// Create the HTTP POST request
	url := fmt.Sprintf("%s/v1/objects/services", client.Hostname)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return nil, err
	}

	// Set the content type
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-HTTP-Method-Override", "GET")

	if client.Username != "" && client.Password != "" {
		req.SetBasicAuth(client.Username, client.Password)
	}

	// Send the request
	resp, err := client.httpClient.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return nil, err
	}

	var responseStruct getServiceResponse
	json.Unmarshal(body, &responseStruct)

	var services []Service
	for _, serviceJson := range responseStruct.Results {
		services = append(services, NewServiceFromJson(serviceJson))
	}

	return services, nil
}

func NewServiceFromJson(serviceJson icinga2ServiceJson) Service {
	service := Service{
		State:     serviceJson.Attributes.State,
		StateType: serviceJson.Attributes.StateType,
	}
	// Split the Name into Hostname and Service name
	parts := strings.SplitN(serviceJson.Name, "!", 2) // Split into at most 2 parts
	if len(parts) > 0 {
		service.HostName = parts[0]
	}
	if len(parts) > 1 {
		service.ServiceName = parts[1]
	}

	return service
}
