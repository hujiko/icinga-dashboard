package main

import (
	"sort"
	"testing"

	"github.com/hujiko/icinga-dashboard/icinga2apiclient"
)

func TestBuildServiceListRecords(t *testing.T) {
	services := []icinga2apiclient.Service{
		{HostName: "host-b", ServiceName: "disk", State: 2, StateType: 1},
		{HostName: "host-a", ServiceName: "disk", State: 2, StateType: 1},
		{HostName: "host-c", ServiceName: "ping", State: 1, StateType: 0},
	}

	records := buildServiceListRecords(services)
	sort.Sort(ByState(records))

	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	aggregated := records[0]
	if aggregated.Name != "disk" {
		t.Fatalf("expected first record to be disk, got %q", aggregated.Name)
	}
	if aggregated.HostField != "2 Hosts" {
		t.Errorf("expected aggregated host field %q, got %q", "2 Hosts", aggregated.HostField)
	}
	if !aggregated.IsAggregated {
		t.Errorf("expected aggregated record to be marked aggregated")
	}
	if aggregated.AggregatedHostsCount != 2 {
		t.Errorf("expected aggregated host count 2, got %d", aggregated.AggregatedHostsCount)
	}
	expectedHosts := []string{"host-b", "host-a"}
	for i, host := range expectedHosts {
		if aggregated.AggregatedHosts[i] != host {
			t.Errorf("expected aggregated host %d to be %q, got %q", i, host, aggregated.AggregatedHosts[i])
		}
	}

	single := records[1]
	if single.Name != "ping" {
		t.Fatalf("expected second record to be ping, got %q", single.Name)
	}
	if single.HostField != "host-c" {
		t.Errorf("expected single host field %q, got %q", "host-c", single.HostField)
	}
	if single.IsAggregated {
		t.Errorf("expected single record not to be aggregated")
	}
	if single.AggregatedHostsCount != 1 {
		t.Errorf("expected single host count 1, got %d", single.AggregatedHostsCount)
	}
	if len(single.AggregatedHosts) != 1 || single.AggregatedHosts[0] != "host-c" {
		t.Errorf("expected single aggregated hosts to contain host-c, got %v", single.AggregatedHosts)
	}
}
