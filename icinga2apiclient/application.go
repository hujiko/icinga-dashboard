package icinga2apiclient

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type IcingaApplication struct {
	EnableEventHandlers bool    `json:"enable_event_handlers"`
	EnableFlapping      bool    `json:"enable_flapping"`
	EnableHostChecks    bool    `json:"enable_host_checks"`
	EnableNotifications bool    `json:"enable_notifications"`
	EnablePerfdata      bool    `json:"enable_perfdata"`
	EnableServiceChecks bool    `json:"enable_service_checks"`
	Environment         string  `json:"environment"`
	NodeName            string  `json:"node_name"`
	PID                 int     `json:"pid"`
	ProgramStart        float64 `json:"program_start"`
	Version             string  `json:"version"`
}

func (client *Client) GetIcingaApplicationStatus() (*IcingaApplication, error) {
	responseBody, err := client.makeRequest(http.MethodGet, "/v1/status/IcingaApplication", nil)
	if err != nil {
		return nil, err
	}

	var raw struct {
		Results []struct {
			Status struct {
				IcingaApplication struct {
					App IcingaApplication `json:"app"`
				} `json:"icingaapplication"`
			} `json:"status"`
		} `json:"results"`
	}

	err = json.Unmarshal(responseBody, &raw)
	if err != nil {
		return nil, err
	}

	if len(raw.Results) == 0 {
		return nil, fmt.Errorf("No results in IcingaApplication response")
	}

	app := raw.Results[0].Status.IcingaApplication.App

	return &app, nil
}
