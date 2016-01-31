// +build linux darwin freebsd

package cmd

import (
	"fmt"
	"github.com/DataDog/datadog-go/statsd"
	"github.com/hpcloud/tail"
	"github.com/spf13/cobra"
	"regexp"
	"strconv"
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
			Log(fmt.Sprintf("line: %s", content), "debug")
			grabTimestamp(content)
		}
		if strings.HasPrefix(content, "queries") {
			Log(fmt.Sprintf("line: %s", content), "debug")
			queriesForwarded(content)
			queriesLocal(content)
			queriesAuthoritativeZones(content)
		}
		if strings.HasPrefix(content, "server") {
			Log(fmt.Sprintf("line: %s", content), "debug")
			serverStats(content)
		}
	}
}

func grabTimestamp(content string) {
	r := regexp.MustCompile(`\d+`)
	timestamp := r.FindString(content)
	unixTimestamp, _ := strconv.Atoi(timestamp)
	Log(fmt.Sprintf("Timestamp: %d", unixTimestamp), "debug")
}

func serverStats(content string) {
	r := regexp.MustCompile(`server (\d+\.\d+\.\d+\.\d+#\d+): queries sent (\d+), retried or failed (\d+)`)
	server := r.FindAllStringSubmatch(content, -1)
	if server != nil {
		srvr := server[0]
		serverAddress := srvr[1]
		serverAddressSent, _ := strconv.Atoi(srvr[2])
		serverAddressRetryFailures, _ := strconv.Atoi(srvr[3])
		Log(fmt.Sprintf("Server: %s Queries: %d Retries/Failures: %d\n", serverAddress, serverAddressSent, serverAddressRetryFailures), "debug")
	}
}

func queriesForwarded(content string) {
	r := regexp.MustCompile(`forwarded (\d+),`)
	forwarded := r.FindAllStringSubmatch(content, -1)
	if forwarded != nil {
		fwd := forwarded[0]
		value := fwd[1]
		queriesForwarded, _ := strconv.Atoi(value)
		Log(fmt.Sprintf("Forwarded Queries: %d", queriesForwarded), "debug")
	}
}

func queriesLocal(content string) {
	r := regexp.MustCompile(`queries answered locally (\d+)`)
	local := r.FindAllStringSubmatch(content, -1)
	if local != nil {
		lcl := local[0]
		lclv := lcl[1]
		localResponses, _ := strconv.Atoi(lclv)
		Log(fmt.Sprintf("Responded Locally: %d", localResponses), "debug")
	}
}

func queriesAuthoritativeZones(content string) {
	r := regexp.MustCompile(`for authoritative zones (\d+)`)
	zones := r.FindAllStringSubmatch(content, -1)
	if zones != nil {
		zone := zones[0]
		zonev := zone[1]
		authoritativeZones, _ := strconv.Atoi(zonev)
		Log(fmt.Sprintf("Authoritative Zones: %d", authoritativeZones), "debug")
	}
}

func dnsmasqSignals() {
	for {
		procs := GetMatches("dnsmasq", false)
		sendUSR1(procs)
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

func sendUSR1(procs []ProcessList) {
	if len(procs) > 0 {
		for _, proc := range procs {
			proc.USR1()
		}
	}
}
