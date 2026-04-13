package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"text/template"
	"time"

	"github.com/hujiko/icinga-dashboard/icinga2apiclient"
)

var (
	client              dashboardClient
	defaultMinState     int
	defaultMaxState     int
	defaultMinStateType int
	baseURL             string
	now                 = time.Now
)

type dashboardClient interface {
	GetIcingaApplicationStatus() (*icinga2apiclient.IcingaApplication, error)
	GetCIBStatus() (*icinga2apiclient.CIBStatus, error)
	GetServices(minState int, maxState int, minStateType int) ([]icinga2apiclient.Service, error)
	GetHosts(minStateType int) ([]icinga2apiclient.Host, error)
}

func main() {
	envVariables := parseEnvVariables()

	apiClient, err := icinga2apiclient.NewClient(
		envVariables["ICINGA2_API_URL"].(string),
		envVariables["ICINGA2_API_CLIENT_CERT_PATH"].(string),
		envVariables["ICINGA2_API_CLIENT_KEY_PATH"].(string),
		envVariables["ICINGA2_API_CA_PATH"].(string),
		envVariables["ICINGA2_API_TIMEOUT"].(int),
		envVariables["ICINGA2_API_VALIDATE_CERTIFICATE"].(int) == 1,
	)
	if err != nil {
		fmt.Printf("Error configuring API client: %v\n", err)
	}

	apiClient.Username = envVariables["ICINGA2_API_USERNAME"].(string)
	apiClient.Password = envVariables["ICINGA2_API_PASSWORD"].(string)
	client = apiClient

	defaultMinState = envVariables["MIN_STATE"].(int)
	defaultMaxState = envVariables["MAX_STATE"].(int)
	defaultMinStateType = envVariables["MIN_STATE_TYPE"].(int)
	baseURL = envVariables["ICINGA2_BASE_URL"].(string)

	// If path starts with /assets/ then serve static files from assets/
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	http.Handle("/favicon.ico", http.FileServer(http.Dir("assets")))

	// Define the handler for the root URL
	http.HandleFunc("/", renderDashboard)
	http.HandleFunc("/api/v1/dashboard", renderJSON)

	fmt.Printf("Starting webserver. Listening on %s\n", envVariables["LISTEN_ADDRESS"])
	err = http.ListenAndServe(envVariables["LISTEN_ADDRESS"].(string), nil)
	if err != nil {
		panic(err) // Handle error if the server fails to start
	}
}

func renderDashboard(w http.ResponseWriter, r *http.Request) {
	pageVariables := buildPageVariables(r)
	if pageVariables.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	tmpl, tmplErr := template.ParseFiles("index.html")
	if tmplErr != nil {
		http.Error(w, tmplErr.Error(), http.StatusInternalServerError)
		return
	}

	errExec := tmpl.Execute(w, pageVariables)
	if errExec != nil {
		http.Error(w, errExec.Error(), http.StatusInternalServerError)
		return
	}
}

func renderJSON(w http.ResponseWriter, r *http.Request) {
	pageVariables := buildPageVariables(r)
	if pageVariables.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(pageVariables)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func buildPageVariables(r *http.Request) PageVariables {
	minStateType := defaultMinStateType
	minState := defaultMinState
	maxState := defaultMaxState

	queryParamters := r.URL.Query()
	if value, err := strconv.Atoi(queryParamters.Get("minStateType")); err == nil {
		minStateType = value
	}
	if value, err := strconv.Atoi(queryParamters.Get("minState")); err == nil {
		minState = value
	}
	if value, err := strconv.Atoi(queryParamters.Get("maxState")); err == nil {
		maxState = value
	}

	currentTime := now()
	pageVariables := PageVariables{
		TimeString:   currentTime.Format("15:04:05"),
		Time:         timestamp{currentTime},
		MinStateType: stateTypeNumToString(minStateType),
		MinState:     stateNumToString(minState),
		MaxState:     stateNumToString(maxState),
		BaseURL:      baseURL,
	}

	if appStatus, err := client.GetIcingaApplicationStatus(); err != nil {
		fmt.Printf("Error getting IcingaApplication status: %v\n", err)
		pageVariables.Error = err
	} else {
		pageVariables.NotificationsDisabled = !appStatus.EnableNotifications
	}

	if cibStatus, err := client.GetCIBStatus(); err != nil {
		fmt.Printf("Error getting hosts: %v\n", err)
		pageVariables.Error = err
	} else {
		pageVariables.CIBStatus = cibStatus
	}

	pageVariables.ServiceRecords = getAndSortServices(client, minState, maxState, minStateType)

	hosts, err := client.GetHosts(minStateType)
	if err != nil {
		fmt.Printf("Error getting hosts: %v\n", err)
		pageVariables.Error = err
	}

	pageVariables.HostRecords = make([]PageHostListRecord, 0)
	for _, host := range hosts {
		pageVariables.HostRecords = append(pageVariables.HostRecords, PageHostListRecord{
			Name:      host.Name,
			State:     host.State,
			StateType: host.StateType,
		})
	}

	sort.Sort(ByState(pageVariables.ServiceRecords))
	sort.Sort(ByName(pageVariables.HostRecords))

	return pageVariables
}

func getAndSortServices(client dashboardClient, minState int, maxState int, minStateType int) []PageServiceListRecord {
	services, err := client.GetServices(minState, maxState, minStateType)
	if err != nil {
		fmt.Printf("Error getting hosts: %v\n", err)
	}

	return buildServiceListRecords(services)
}

func buildServiceListRecords(services []icinga2apiclient.Service) []PageServiceListRecord {
	groupedServices := make(map[string][]icinga2apiclient.Service)
	for _, service := range services {
		groupKey := fmt.Sprintf("%s-%d-%d", service.ServiceName, service.State, service.StateType)
		groupedServices[groupKey] = append(groupedServices[groupKey], service)
	}

	var resultSet []PageServiceListRecord

	for _, group := range groupedServices {
		var hostField string
		var isAggregated bool
		if len(group) == 1 {
			hostField = group[0].HostName
			isAggregated = false
		} else {
			hostField = fmt.Sprintf("%d Hosts", len(group))
			isAggregated = true
		}

		resultSet = append(resultSet, PageServiceListRecord{
			HostField:    hostField,
			Name:         group[0].ServiceName,
			State:        group[0].State,
			StateType:    group[0].StateType,
			IsAggregated: isAggregated,
			AggregatedHosts: func() []string {
				var hosts []string
				for _, service := range group {
					hosts = append(hosts, service.HostName)
				}
				return hosts
			}(),
			AggregatedHostsCount: len(group),
		})
	}

	return resultSet
}

func parseEnvVariables() map[string]interface{} {
	// Define the environment variable names and their default values
	varDefaults := map[string]interface{}{
		// Listen Address
		// The IP and Port you want the HTTP-Server to bind to.
		"LISTEN_ADDRESS": ":8080",

		// Base url of your icinga instance.
		// Used when linking from the dashboard towards icinga.
		// Usually that is: https://icinga.example.com/monitoring
		// Don't add a trailing slash here!
		"ICINGA2_BASE_URL": "",

		// URL of the Icinga2 REST API.
		// Usually that is: https://icinga.example.com:5665/
		// Don't add a trailing slash here!
		"ICINGA2_API_URL": "",

		// Timeout in seconds for HTTP requests towards the API.
		"ICINGA2_API_TIMEOUT": 5,

		// In case your icinga2 API uses client certificates for authentication
		// specify the path of the client private key here.
		"ICINGA2_API_CLIENT_KEY_PATH": "",

		// In case your icinga2 API uses client certificates for authentication
		// specify the path of the client certificate here.
		"ICINGA2_API_CLIENT_CERT_PATH": "",

		// In case you use a self-signed SSL Certificate for the Icinga2-API
		// Specify the path of the root CA here that will be used for validation.
		"ICINGA2_API_CA_PATH": "",

		// In case you use a self-signed SSL Certificate for the Icinga2-API
		// but don't want to validate certificates at all, set this.
		// Possible values
		// 0 => Do not validate
		// 1 => Validate certificate
		"ICINGA2_API_VALIDATE_CERTIFICATE": 1,

		// If you use don't use certificate based authentication, but rely on username and password
		// Define the username here
		"ICINGA2_API_USERNAME": "",

		// If you use don't use certificate based authentication, but rely on username and password
		// Define the password here
		"ICINGA2_API_PASSWORD": "",

		// Lowest state a service has to be in, in order to be shown on the dashboard.
		// Possible values
		// 0 => OK
		// 1 => Warning
		// 2 => Critical
		// 3 => Unknown
		// This value can be overwritten by the query parameter "minState" when opening the dashboard in a browser.
		"MIN_STATE": 1,

		// Highest state a service/host can be in, in order to be shown on the dashboard.
		// Possible values
		// 0 => OK
		// 1 => Warning
		// 2 => Critical
		// 3 => Unknown
		// This value can be overwritten by the query parameter "maxState" when opening the dashboard in a browser.
		"MAX_STATE": 2,

		// min state type a service/host should have, in order to be shown on the dashboard.
		// Possible values
		// 0 => Soft state
		// 1 => hard state
		// This value can be overwritten by the query parameter "maxStateType" when opening the dashboard in a browser.
		"MIN_STATE_TYPE": 0,
	}

	// Create a map to store the retrieved values
	envValues := make(map[string]interface{})

	// Loop through the variable names and get their values
	for varName, defaultValue := range varDefaults {
		value := os.Getenv(varName)
		if value == "" {
			// If the environment variable is not set, use the default value
			envValues[varName] = defaultValue
		} else {
			// If the variable is set, check if it needs to be parsed as an int
			if _, ok := defaultValue.(int); ok {
				if intValue, err := strconv.Atoi(value); err == nil {
					envValues[varName] = intValue
				} else {
					fmt.Printf("Error parsing %s: %s (using default value)\n", varName, value)
					envValues[varName] = defaultValue
				}
			} else {
				envValues[varName] = value
			}
		}
	}

	requiredVars := []string{"LISTEN_ADDRESS", "ICINGA2_BASE_URL", "ICINGA2_API_URL"}
	for _, varName := range requiredVars {
		if envValues[varName] == "" {
			panic(varName + " can't be empty!")
		}
	}

	return envValues
}

func stateNumToString(state int) string {
	mapping := []string{
		"OK",
		"Warning",
		"Critical",
		"Unknown",
	}

	if state < 0 || state >= len(mapping) {
		return "---"
	}

	return mapping[state]
}

func stateTypeNumToString(stateType int) string {
	mapping := []string{
		"Soft",
		"Hard",
	}

	if stateType < 0 || stateType >= len(mapping) {
		return "---"
	}

	return mapping[stateType]
}
