// +build linux darwin freebsd

package cmd

import (
	"fmt"
	"github.com/cloudfoundry/gosigar"
	_ "github.com/davecgh/go-spew/spew" // I want to use this sometimes.
	"strings"
)

// ProcessList is a simplified list of processes on a system.
type ProcessList struct {
	Pname string
	Pid   int
	Pmem  uint64 // in K
}

// GetMatches returns only the matches we want from running processes.
func GetMatches(match string) []ProcessList {
	var Matches []ProcessList
	processes := GetProcessList()
	// spew.Dump(processes)
	Matches = MatchProcessList(*processes, match)
	return Matches
}

// GetProcessList returns all the processes.
func GetProcessList() *[]ProcessList {
	pids := GetPIDs()
	processes := ConvertProcessList(pids)
	return processes
}

// GetPIDs returns a pointer to all pids on machine.
func GetPIDs() *sigar.ProcList {
	pids := sigar.ProcList{}
	pids.Get()
	return &pids
}

// ConvertProcessList converts the *sigar.ProcList into our []ProcessList struct.
func ConvertProcessList(p *sigar.ProcList) *[]ProcessList {
	var List []ProcessList
	var proc ProcessList
	for _, pid := range p.List {
		var memory uint64
		state := sigar.ProcState{}
		mem := sigar.ProcMem{}
		time := sigar.ProcTime{}
		if err := state.Get(pid); err != nil {
			continue
		}
		if err := mem.Get(pid); err != nil {
			continue
		}
		if err := time.Get(pid); err != nil {
			continue
		}
		memory = mem.Resident
		proc = ProcessList{Pname: state.Name, Pid: pid, Pmem: memory}
		List = append(List, proc)
	}
	return &List
}

// MatchProcessList looks through the struct processes that match.
func MatchProcessList(procs []ProcessList, match string) []ProcessList {
	var Matches []ProcessList
	for _, proc := range procs {
		if proc.Pname == match || proc.Pname == "goshe" {
			Matches = append(Matches, proc)
		}
	}
	return Matches
}

// SendMetrics sends memory metrics to Dogstatsd.
func SendMetrics(p []ProcessList) bool {
	var err error
	dog := DogConnect()
	for _, proc := range p {
		processName := strings.ToLower(strings.Replace(proc.Pname, " ", "_", -1))
		metricName := fmt.Sprintf("%s.rss_memory", processName)
		Log(fmt.Sprintf("SendMetrics process='%#v' processName='%s' metricName='%s' memory='%b'", proc, processName, metricName, float64(proc.Pmem)), "debug")
		err = dog.Histogram(metricName, float64(proc.Pmem), dog.Tags, 1)
		if err != nil {
			Log(fmt.Sprintf("Error sending rss_memory stats for '%s'", processName), "info")
			return false
		}
	}
	return true
}
