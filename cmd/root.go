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
	Short: "Additional Apache memory stats to datadog.",
	Long:  `Additional Apache memory stats to datadog.`,
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
)

func init() {
	Direction = SetDirection()
	RootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "", false, "log output to stdout")
	RootCmd.PersistentFlags().StringVarP(&ProcessName, "process", "p", "", "Process name to match.")
}
