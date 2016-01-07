// +build linux darwin freebsd

package main

import (
	"./cmd/"
	"fmt"
	"log"
	"log/syslog"
	"os"
	"runtime"
)

// CompileDate tracks when the binary was compiled. It's inserted during a build
// with build flags. Take a look at the Makefile for information.
var CompileDate = "No date provided."

// GitCommit tracks the SHA of the built binary. It's inserted during a build
// with build flags. Take a look at the Makefile for information.
var GitCommit = "No revision provided."

// Version is the version of the built binary. It's inserted during a build
// with build flags. Take a look at the Makefile for information.
var Version = "No version provided."

// GoVersion details the version of Go this was compiled with.
var GoVersion = runtime.Version()

// The name of the program.
var programName = "goshe"

func main() {
	logwriter, e := syslog.New(syslog.LOG_NOTICE, programName)
	if e == nil {
		log.SetOutput(logwriter)
	}
	cmd.Log(fmt.Sprintf("%s version: %s", programName, Version), "info")

	args := os.Args[1:]
	for _, arg := range args {
		if arg == "-v" || arg == "--version" {
			fmt.Printf("Version  : %s\nRevision : %s\nDate     : %s\nGo       : %s\n", Version, GitCommit, CompileDate, GoVersion)
			os.Exit(0)
		}
	}

	runtime.GOMAXPROCS(runtime.NumCPU())
	cmd.RootCmd.Execute()
}
