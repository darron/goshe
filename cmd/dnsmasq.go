// +build linux darwin freebsd

package cmd

import (
	"fmt"
	"github.com/DataDog/datadog-go/statsd"
	"github.com/hpcloud/tail"
	"github.com/spf13/cobra"
	"strings"
	"time"
)

const (
	dnsmasqLog     = "/var/log/dnsmasq/dnsmasq"
	signalInterval = 60
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
		dnsmasqSignalStats(t, dog)
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
	FullLogs bool
)

func init() {
	dnsmasqCmd.Flags().StringVarP(&DnsmasqLog, "log", "", dnsmasqLog, "dnsmasq log file.")
	dnsmasqCmd.Flags().BoolVarP(&FullLogs, "full", "", false, "Use full --log-queries logs.")
	RootCmd.AddCommand(dnsmasqCmd)
}

// SendLineStats sends the stats to Datadog.
func SendLineStats(dog *statsd.Client, line string, metric string) {
	Log(fmt.Sprintf("%s: %s", metric, line), "debug")
	oldTags := dog.Tags
	dog.Tags = append(dog.Tags, fmt.Sprintf("record:%s", metric))
	dog.Count("dnsmasq.event", 1, dog.Tags, 1)
	dog.Tags = oldTags
}

// Example Logs:
// Jan 29 20:32:55 dnsmasq[29389]: time 1454099575
// Jan 29 20:32:55 dnsmasq[29389]: cache size 150, 41/1841 cache insertions re-used unexpired cache entries.
// Jan 29 20:32:55 dnsmasq[29389]: queries forwarded 354453, queries answered locally 251099667
// Jan 29 20:32:55 dnsmasq[29389]: server 127.0.0.1#8600: queries sent 142940, retried or failed 0
// Jan 29 20:32:55 dnsmasq[29389]: server 172.16.0.23#53: queries sent 211510, retried or failed 0

func dnsmasqSignalStats(t *tail.Tail, dog *statsd.Client) {
	go dnsmasqSignals()
	for line := range t.Lines {
		// Blank lines really mess this up - this protects against it.
		if line.Text == "" {
			continue
		}
		content := strings.Split(line.Text, "]: ")[1]
		if strings.HasPrefix(content, "time") {
			fmt.Printf("time: %s\n", content)
		}
		if strings.HasPrefix(content, "queries") {
			fmt.Printf("queries: %s\n", content)
		}
		if strings.HasPrefix(content, "server") {
			fmt.Printf("server: %s\n", content)
		}
	}
}

func dnsmasqSignals() {
	var procs []ProcessList
	for {
		procs = GetMatches("dnsmasq", false)
		if len(procs) >= 1 {
			for _, proc := range procs {
				proc.USR1()
			}
		}
		time.Sleep(time.Duration(signalInterval) * time.Second)
	}
}

func dnsmasqFullLogsStats(t *tail.Tail, dog *statsd.Client) {
	for line := range t.Lines {
		content := strings.Split(line.Text, "]: ")[1]
		if strings.HasPrefix(content, "/") {
			SendLineStats(dog, content, "hosts")
			continue
		}
		if strings.HasPrefix(content, "query") {
			SendLineStats(dog, content, "query")
			continue
		}
		if strings.HasPrefix(content, "cached") {
			SendLineStats(dog, content, "cached")
			continue
		}
		if strings.HasPrefix(content, "forwarded") {
			SendLineStats(dog, content, "forwarded")
			continue
		}
		if strings.HasPrefix(content, "reply") {
			SendLineStats(dog, content, "reply")
			continue
		}
	}
}
