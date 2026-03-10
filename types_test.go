package main

import (
	"net/url"
	"sort"
	"testing"
)

func TestPageHostListRecord_URLEncodedHost(t *testing.T) {
	r := PageHostListRecord{Name: "host name"}
	expected := url.QueryEscape("host name")
	if r.URLEncodedHost() != expected {
		t.Errorf("URLEncodedHost = %v, want %v", r.URLEncodedHost(), expected)
	}
}

func TestPageServiceListRecord_URLEncodedHost(t *testing.T) {
	r := PageServiceListRecord{HostField: "host field"}
	expected := url.QueryEscape("host field")
	if r.URLEncodedHost() != expected {
		t.Errorf("URLEncodedHost = %v, want %v", r.URLEncodedHost(), expected)
	}
}

func TestPageServiceListRecord_URLEncodedService(t *testing.T) {
	r := PageServiceListRecord{Name: "service name"}
	expected := url.QueryEscape("service name")
	if r.URLEncodedService() != expected {
		t.Errorf("URLEncodedService = %v, want %v", r.URLEncodedService(), expected)
	}
}

func TestByState_Sort(t *testing.T) {
	list := ByState{
		{Name: "A", State: 2, StateType: 1},
		{Name: "B", State: 1, StateType: 0},
		{Name: "C", State: 2, StateType: 0},
	}
	// sort
	sort.Slice(list, func(i, j int) bool { return list.Less(i, j) })
	// Expected order: A (2,1), C (2,0), B (1,0)
	expected := []string{"A", "C", "B"}
	for i, name := range expected {
		if list[i].Name != name {
			t.Errorf("Sort failed at %d: got %v, want %v", i, list[i].Name, name)
		}
	}
}

func TestByName_Sort(t *testing.T) {
	list := ByName{
		{Name: "B"},
		{Name: "A"},
	}
	list.Swap(0, 1)
	if list[0].Name != "A" {
		t.Errorf("Swap failed: got %v", list[0].Name)
	}
	if !list.Less(0, 1) {
		t.Errorf("Less failed: expected true")
	}
}

func TestTimestamp_MarshalUnmarshalJSON(t *testing.T) {
	ts := timestamp{}
	data, err := ts.MarshalJSON()
	if err != nil {
		t.Errorf("MarshalJSON error: %v", err)
	}
	var ts2 timestamp
	err = ts2.UnmarshalJSON(data)
	if err != nil {
		t.Errorf("UnmarshalJSON error: %v", err)
	}
}
