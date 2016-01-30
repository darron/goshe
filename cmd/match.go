// +build linux darwin freebsd

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var matchCmd = &cobra.Command{
	Use:   "match",
	Short: "Grab memory stats from matching processes and sends to Datadog.",
	Long:  `Grab memory stats from matching processes and sends to Datadog.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		checkMatchFlags()
	},
	Run: startMatch,
}

func startMatch(cmd *cobra.Command, args []string) {
	for {
		matches := GetMatches(ProcessName, true)
		if matches != nil {
			fmt.Printf("Found %d matches.\n", len(matches)-1)
			SendMetrics(matches)
		} else {
			fmt.Printf("Did not find any matches.\n")
		}
		time.Sleep(time.Duration(Interval) * time.Second)
	}
}

func checkMatchFlags() {
	if ProcessName == "" {
		fmt.Println("Need to specify the process to search for: -p")
		os.Exit(1)
	}
	fmt.Println("Press CTRL-C to shutdown.")
}

func init() {
	RootCmd.AddCommand(matchCmd)
}
