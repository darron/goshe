// +build linux darwin freebsd

package main

import (
	_ "expvar"
	"fmt"
	"github.com/Datadog/goshe/cmd"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
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
	// Setup nice shutdown with CTRL-C.
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go handleCtrlC(c)

	// Listen for expvar if we have GOSHE_DEBUG set.
	if os.Getenv("GOSHE_DEBUG") != "" {
		go setupExpvarHTTP()
	}

	cmd.RootCmd.Execute()
}

func setupExpvarHTTP() {
	// Listen for expvar
	http.ListenAndServe(":1313", nil)
}

// Any cleanup tasks on shutdown could happen here.
func handleCtrlC(c chan os.Signal) {
	sig := <-c
	message := fmt.Sprintf("Received '%s' - shutting down.", sig)
	cmd.Log(message, "info")
	fmt.Printf("%s\n", message)
	os.Exit(0)
}
