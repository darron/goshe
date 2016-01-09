// +build linux darwin freebsd

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"time"
)

var apacheCmd = &cobra.Command{
	Use:   "apache",
	Short: "Grab stats from matching Apache2 processes and sends to Datadog.",
	Long:  `Grab stats from matching Apache2 processes and sends to Datadog.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		checkApacheFlags()
	},
	Run: startApache,
}

func startApache(cmd *cobra.Command, args []string) {
	for {
		matches := GetMatches(ProcessName)
		if matches != nil {
			fmt.Printf("Found %d matches.\n", len(matches)-1)
			SendMetrics(matches)
		} else {
			fmt.Printf("Did not find any matches.\n")
		}
		time.Sleep(time.Duration(Interval) * time.Second)
	}
}

func checkApacheFlags() {
	if ProcessName == "" {
		ProcessName = "apache2"
	}
	fmt.Println("Press CTRL-C to shutdown.")
}

func init() {
	RootCmd.AddCommand(apacheCmd)
}
