package cmd

import (
	"testing"
)

func TestGetPIDs(t *testing.T) {
	pids := GetPIDs()
	if pids == nil {
		t.Error("Got a nil *sigar.ProcList.")
	}
}

func TestConvertProcessList(t *testing.T) {
	pids := GetPIDs()
	processes := ConvertProcessList(pids)
	if processes == nil {
		t.Error("Got a nil *[]ProcessList.")
	}
}

func TestMatchProcessList(t *testing.T) {
	match := "go"
	pids := GetPIDs()
	processes := ConvertProcessList(pids)
	matches := MatchProcessList(*processes, match)
	if matches == nil {
		t.Error("Got nill matches")
	}
}

func TestGetProcessList(t *testing.T) {
	processes := GetProcessList()
	if processes == nil {
		t.Error("Didn't get any processes.")
	}
}

func TestGetMatches(t *testing.T) {
	match := "go"
	matches := GetMatches(match)
	if matches == nil {
		t.Error("Got no matches.")
	}
}

func TestSendMetrics(t *testing.T) {
	matches := GetMatches("go")
	if matches != nil {
		result := SendMetrics(matches)
		if !result {
			t.Error("Did not send Go RSS Metrics.")
		}
	} else {
		t.Error("Didn't find any matches.")
	}
}
