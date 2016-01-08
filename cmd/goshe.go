// +build linux darwin freebsd

package cmd

import (
	"github.com/cloudfoundry/gosigar"
	_ "github.com/davecgh/go-spew/spew" // I want to use this sometimes.
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
		state := sigar.ProcState{}
		mem := sigar.ProcMem{}
		if err := state.Get(pid); err != nil {
			continue
		}
		if err := mem.Get(pid); err != nil {
			continue
		}
		proc = ProcessList{Pname: state.Name, Pid: pid, Pmem: mem.Resident / 1024}
		List = append(List, proc)
	}
	return &List
}

// MatchProcessList looks through the struct processes that match.
func MatchProcessList(procs []ProcessList, match string) []ProcessList {
	var Matches []ProcessList
	for _, proc := range procs {
		if proc.Pname == match {
			Matches = append(Matches, proc)
		}
	}
	return Matches
}
