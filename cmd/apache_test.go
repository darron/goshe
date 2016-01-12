package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetServerStatus(t *testing.T) {
	ts := httptest.NewServer(http.FileServer(http.Dir("../test/http")))
	url := fmt.Sprintf("%s/index.html", ts.URL)
	serverStatus := getServerStatus(url)
	if serverStatus == nil {
		t.Error("That's not good! Not getting html.Nodes back.")
	}
}

func TestParseServerStatus(t *testing.T) {
	ts := httptest.NewServer(http.FileServer(http.Dir("../test/http")))
	url := fmt.Sprintf("%s/index.html", ts.URL)
	serverStatus := getServerStatus(url)
	stringResults := parseServerStatus(serverStatus)
	if len(stringResults) != 64 {
		t.Error("We're not getting the right number of strings back. Should be 64.")
	}
}

func TestParseServerStats(t *testing.T) {
	ts := httptest.NewServer(http.FileServer(http.Dir("../test/http")))
	url := fmt.Sprintf("%s/index.html", ts.URL)
	serverStatus := getServerStatus(url)
	stringResults := parseServerStatus(serverStatus)
	ApacheProcesses := parseProcessStats(stringResults)
	if len(ApacheProcesses) != 9 {
		t.Error("That's bad - we should see 9 Apache structs.")
	}
}

func createTestProcessList() []ProcessList {
	var procs []ProcessList
	var proc ProcessList
	proc = ProcessList{Pname: "apache2", Pid: 10434, Pmem: 10520}
	procs = append(procs, proc)
	proc = ProcessList{Pname: "apache2", Pid: 10360, Pmem: 20520}
	procs = append(procs, proc)
	proc = ProcessList{Pname: "apache2", Pid: 10282, Pmem: 30520}
	procs = append(procs, proc)
	proc = ProcessList{Pname: "apache2", Pid: 10345, Pmem: 15520}
	procs = append(procs, proc)
	proc = ProcessList{Pname: "apache2", Pid: 10475, Pmem: 25520}
	procs = append(procs, proc)
	return procs
}

func TestCreateProcessMemMap(t *testing.T) {
	procs := createTestProcessList()
	processMap := createProcessMemMap(procs)
	if processMap[10434] != uint64(10520) {
		t.Error("That's incorrect - it should be uint64(10520).")
	}
}

func createTestApacheList() []ApacheProcess {
	var stats []ApacheProcess
	var apache ApacheProcess
	apache = ApacheProcess{Pid: 10434, Vhost: "andy.bam.nonwebdev.com"}
	stats = append(stats, apache)
	apache = ApacheProcess{Pid: 10360, Vhost: "jon.bam.nonwebdev.com"}
	stats = append(stats, apache)
	apache = ApacheProcess{Pid: 10282, Vhost: "darron.bam.nonwebdev.com"}
	stats = append(stats, apache)
	apache = ApacheProcess{Pid: 10345, Vhost: "darron.bam.nonwebdev.com"}
	stats = append(stats, apache)
	apache = ApacheProcess{Pid: 10475, Vhost: "robb.bam.nonwebdev.com"}
	stats = append(stats, apache)
	return stats
}

// Testing to see if the stats get sent.
func TestSendApacheServerStats(t *testing.T) {
	procs := createTestProcessList()
	if len(procs) != 5 {
		t.Error("That's a problem - there should be 5 processes.")
	}
	apaches := createTestApacheList()
	if len(apaches) != 5 {
		t.Error("That's a problem - there should be 5 Apaches.")
	}
	procMap := createProcessMemMap(procs)
	SendApacheServerStats(apaches, procMap)
}

// Adding an extra Apache process without corresponding memory info.
func TestSendApacheServerStatsWithExtraApache(t *testing.T) {
	procs := createTestProcessList()
	apaches := createTestApacheList()
	apache := ApacheProcess{Pid: 10800, Vhost: "wildcard.bam.nonwebdev.com"}
	apaches = append(apaches, apache)
	procMap := createProcessMemMap(procs)
	SendApacheServerStats(apaches, procMap)
}

// Adding an extra Process without a matching Apache process.
func TestSendApacheServerStatsWithExtraProcess(t *testing.T) {
	procs := createTestProcessList()
	proc := ProcessList{Pname: "apache2", Pid: 16475, Pmem: 255520}
	procs = append(procs, proc)
	apaches := createTestApacheList()
	procMap := createProcessMemMap(procs)
	SendApacheServerStats(apaches, procMap)
}
