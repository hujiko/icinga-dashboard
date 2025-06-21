package main

import (
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
	client              *icinga2apiclient.Client
	defaultMinState     int
	defaultMaxState     int
	defaultMinStateType int
	baseURL             string
)

func main() {
	envVariables := parseEnvVariables()

	var err error
	client, err = icinga2apiclient.NewClient(
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

	client.Username = envVariables["ICINGA2_API_USERNAME"].(string)
	client.Password = envVariables["ICINGA2_API_PASSWORD"].(string)

	defaultMinState = envVariables["MIN_STATE"].(int)
	defaultMaxState = envVariables["MAX_STATE"].(int)
	defaultMinStateType = envVariables["MIN_STATE_TYPE"].(int)
	baseURL = envVariables["ICINGA2_BASE_URL"].(string)

	// If path starts with /assets/ then serve static files from assets/
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	http.Handle("/favicon.ico", http.FileServer(http.Dir("assets")))

	// Define the handler for the root URL
	http.HandleFunc("/", renderDashboard)

	fmt.Printf("Starting webserver. Listening on %s\n", envVariables["LISTEN_ADDRESS"])
	err = http.ListenAndServe(envVariables["LISTEN_ADDRESS"].(string), nil)
	if err != nil {
		panic(err) // Handle error if the server fails to start
	}
}

func renderDashboard(w http.ResponseWriter, r *http.Request) {
	minStateType := defaultMinStateType
	minState := defaultMinState
	maxState := defaultMaxState

	// Parse the HTML template
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

	// Define the data to pass to the template
	pageVariables := PageVariables{
		Time:         time.Now().Format("15:04:05"),
		MinStateType: stateTypeNumToString(minStateType),
		MinState:     stateNumToString(minState),
		MaxState:     stateNumToString(maxState),
		BaseURL:      baseURL,
	}

	cibStatus, err := client.GetCIBStatus()
	if err != nil {
		fmt.Printf("Error getting hosts: %v\n", err)
		pageVariables.Error = err
		err = tmpl.Execute(w, pageVariables)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	pageVariables.CIBStatus = cibStatus

	pageVariables.ServiceRecords = getAndSortServices(client, minState, maxState, minStateType)

	hosts, err := client.GetHosts(minStateType)
	if err != nil {
		fmt.Printf("Error getting hosts: %v\n", err)
		pageVariables.Error = err
		err = tmpl.Execute(w, pageVariables)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	var hostList []PageHostListRecord
	for _, host := range hosts {
		hostList = append(hostList, PageHostListRecord{
			Name:      host.Name,
			State:     host.State,
			StateType: host.StateType,
		})
	}
	pageVariables.HostRecords = hostList

	// Sort the services by State
	sort.Sort(ByState(pageVariables.ServiceRecords))

	// Sort the hosts by Name
	sort.Sort(ByName(pageVariables.HostRecords))

	// Execute the template with the data
	err = tmpl.Execute(w, pageVariables)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getAndSortServices(client *icinga2apiclient.Client, minState int, maxState int, minStateType int) []PageServiceListRecord {
	services, err := client.GetServices(minState, maxState, minStateType)
	if err != nil {
		fmt.Printf("Error getting hosts: %v\n", err)
	}

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
			ServiceField: group[0].ServiceName,
			State:        group[0].State,
			StateType:    group[0].StateType,
			IsAggregated: isAggregated,
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

	if state > len(mapping) {
		return "---"
	}

	return mapping[state]
}

func stateTypeNumToString(stateType int) string {
	mapping := []string{
		"Soft",
		"Hard",
	}

	if stateType > len(mapping) {
		return "---"
	}

	return mapping[stateType]
}
