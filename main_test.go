package main

import (
	"testing"
)

func TestStateNumToString(t *testing.T) {
	cases := []struct {
		input    int
		expected string
	}{
		{0, "OK"},
		{1, "Warning"},
		{2, "Critical"},
		{3, "Unknown"},
		{4, "---"},
	}
	for _, c := range cases {
		result := stateNumToString(c.input)
		if result != c.expected {
			t.Errorf("stateNumToString(%d) = %v, want %v", c.input, result, c.expected)
		}
	}
}

func TestStateTypeNumToString(t *testing.T) {
	cases := []struct {
		input    int
		expected string
	}{
		{0, "Soft"},
		{1, "Hard"},
		{2, "---"},
	}
	for _, c := range cases {
		result := stateTypeNumToString(c.input)
		if result != c.expected {
			t.Errorf("stateTypeNumToString(%d) = %v, want %v", c.input, result, c.expected)
		}
	}
}
