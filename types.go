package main

import (
	"net/url"

	"github.com/hujiko/icinga-dashboard/icinga2apiclient"
)

type PageVariables struct {
	Title          string
	Time           string
	ServiceRecords []PageServiceListRecord
	HostRecords    []PageHostListRecord
	CIBStatus      *icinga2apiclient.CIBStatus
	MinStateType   string
	MinState       string
	MaxState       string
	Error          error
	BaseURL        string
}

type PageServiceListRecord struct {
	HostField    string
	ServiceField string
	State        int
	StateType    int
	Name         string
	IsAggregated bool
}

type PageHostListRecord struct {
	State     int
	StateType int
	Name      string
}

func (r *PageHostListRecord) URLEncodedHost() string {
	return url.QueryEscape(r.Name)
}

func (r *PageServiceListRecord) URLEncodedHost() string {
	return url.QueryEscape(r.HostField)
}

func (r *PageServiceListRecord) URLEncodedService() string {
	return url.QueryEscape(r.ServiceField)
}

// So that Services can be sorted by State
type ByState []PageServiceListRecord

func (a ByState) Len() int { return len(a) }
func (a ByState) Less(i, j int) bool {
	// Sort by State descending
	if a[i].State != a[j].State {
		return a[i].State > a[j].State // Descending order
	}
	// If States are equal, sort by ServiceName ascending
	if a[i].ServiceField != a[j].ServiceField {
		return a[i].ServiceField < a[j].ServiceField
	}

	return a[i].StateType < a[j].StateType // Ascending order
}
func (a ByState) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// So that Hosts can be sorted by Name
type ByName []PageHostListRecord

func (a ByName) Len() int { return len(a) }
func (a ByName) Less(i, j int) bool {
	return a[i].Name < a[j].Name // Ascending order
}
func (a ByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
