package icinga2apiclient

import (
	"encoding/json"
	"fmt"
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

	responseBody, err := client.makeRequest(http.MethodPost, "/v1/objects/services", jsonPayload)
	if err != nil {
		fmt.Printf("Error fetching services: %v\n", err)
		return nil, err
	}

	var responseStruct getServiceResponse
	err = json.Unmarshal(responseBody, &responseStruct)
	if err != nil {
		fmt.Printf("Unable to parse JSON: %v\n", err)
		return nil, err
	}

	var services []Service
	for _, serviceJSON := range responseStruct.Results {
		services = append(services, NewServiceFromJSON(serviceJSON))
	}

	return services, nil
}

func NewServiceFromJSON(serviceJSON icinga2serviceJSON) Service {
	service := Service{
		State:     serviceJSON.Attributes.State,
		StateType: serviceJSON.Attributes.StateType,
	}
	// Split the Name into Hostname and Service name
	parts := strings.SplitN(serviceJSON.Name, "!", 2) // Split into at most 2 parts
	if len(parts) > 0 {
		service.HostName = parts[0]
	}
	if len(parts) > 1 {
		service.ServiceName = parts[1]
	}

	return service
}
