// +build linux darwin freebsd

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
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
	matches := GetMatches(ProcessName)
	if matches != nil {
		fmt.Printf("Found %d matches.\n", len(matches))
	} else {
		fmt.Printf("Did not find any matches.\n")
	}
}

func checkApacheFlags() {
	if ProcessName == "" {
		ProcessName = "apache2"
	}
}

func init() {
	RootCmd.AddCommand(apacheCmd)
}
