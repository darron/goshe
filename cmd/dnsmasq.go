// +build linux darwin freebsd

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

const (
	dnsmasqLog      = "/var/log/dnsmasq/dnsmasq"
	signalInterval  = 20
	yearSetInterval = 1
)

var dnsmasqCmd = &cobra.Command{
	Use:   "dnsmasq",
	Short: "Grab stats from dnsmasq logs and send to Datadog.",
	Long:  `Grab stats from dnsmasq logs and send to Datadog.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		checkDnsmasqFlags()
	},
	Run: startDnsmasq,
}

func startDnsmasq(cmd *cobra.Command, args []string) {
	dog := DogConnect()
	t := OpenLogfile(DnsmasqLog)
	if FullLogs {
		dnsmasqFullLogsStats(t, dog)
	} else {
		dnsmasqSignalStats(t)
	}
}

func checkDnsmasqFlags() {
	fmt.Println("Press CTRL-C to shutdown.")
}

var (
	// DnsmasqLog is the logfile that dnsmasq logs to.
	DnsmasqLog string

	// FullLogs determines whether we're looking at '--log-queries'
	// levels of logs for dnsmasq.
	// It's disabled by default as it's pretty inefficient.
	FullLogs bool

	// CurrentTimestamp is the current timestamp from the dnsmasq logs.
	CurrentTimestamp int64

	// StatsCurrent is the current timestamp's stats.
	StatsCurrent *DNSStats

	// StatsPrevious is the last timestamp's stats.
	StatsPrevious *DNSStats

	// MemoryFlag sends dnsmasq memory stats to Datadog if true.
	MemoryFlag bool
)

func init() {
	dnsmasqCmd.Flags().StringVarP(&DnsmasqLog, "log", "", dnsmasqLog, "dnsmasq log file.")
	dnsmasqCmd.Flags().BoolVarP(&FullLogs, "full", "", false, "Use full --log-queries logs.")
	dnsmasqCmd.Flags().BoolVarP(&MemoryFlag, "mem", "", false, "Send dnsmasq memory stats.")
	RootCmd.AddCommand(dnsmasqCmd)
}
