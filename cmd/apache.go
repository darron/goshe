// +build linux darwin freebsd

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"
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
		SendApacheRSSMetrics(matches)
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

// SendApacheRSSMetrics sends metrics to Dogstatsd.
func SendApacheRSSMetrics(p []ProcessList) bool {
	var err error
	processName := strings.ToLower(strings.Replace(ProcessName, " ", "_", -1))
	metricName := fmt.Sprintf("%s.rss_memory", processName)
	dog := DogConnect()
	for _, proc := range p {
		err = dog.Histogram(metricName, float64(proc.Pmem), dog.Tags, 1)
		if err != nil {
			Log(fmt.Sprintf("Error sending rss_memory stats for '%s'", ProcessName), "info")
			return false
		}
	}
	return true
}
