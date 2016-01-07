package cmd

import (
	"runtime"
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
	match := "getty"
	pids := GetPIDs()
	processes := ConvertProcessList(pids)
	if runtime.GOOS == "darwin" {
		match = "pboard"
	}
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
	match := "getty"
	if runtime.GOOS == "darwin" {
		match = "pboard"
	}
	matches := GetMatches(match)
	if matches == nil {
		t.Error("Got no matches.")
	}
}
