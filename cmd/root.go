// +build linux darwin freebsd

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// RootCmd is the default Cobra struct that starts it all off.
// https://github.com/spf13/cobra
var RootCmd = &cobra.Command{
	Use:   "goshe",
	Short: "Additional stats to datadog.",
	Long:  `Additional stats to datadog.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("`goshe -h` for help information.")
		fmt.Println("`goshe -v` for version information.")
	},
}

var (
	// Direction adds information about which command is running to the logs.
	Direction string

	// Verbose logs all output to stdout.
	Verbose bool

	// ProcessName is the process to match.
	ProcessName string

	// MetricPrefix prefixes all metrics emitted.
	MetricPrefix string

	// Interval is the amount of seconds to loop.
	Interval int
)

func init() {
	Direction = SetDirection()
	LoadConfig()
	RootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "", false, "log output to stdout")
	RootCmd.PersistentFlags().StringVarP(&ProcessName, "process", "p", "", "Process name to match.")
	RootCmd.PersistentFlags().StringVarP(&MetricPrefix, "prefix", "", "goshe", "Metric prefix.")
	RootCmd.PersistentFlags().IntVarP(&Interval, "interval", "i", 5, "Interval when running in a loop.")
}
