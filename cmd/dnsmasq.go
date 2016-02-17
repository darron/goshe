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
	dnsmasqLog      = "/var/log/dnsmasq/dnsmasq"
	signalInterval  = 60
	yearSetInterval = 10
)

// DNSServer is data gathered from a dnsmasq server log line.
type DNSServer struct {
	timestamp     int64
	address       string
	queriesSent   int64
	queriesFailed int64
}

// DNSStats is data gathered from dnsmasq time, queries and server lines.
type DNSStats struct {
	timestamp          int64
	queriesForwarded   int64
	queriesLocal       int64
	authoritativeZones int64
	servers            []DNSServer
}

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

	// CurrentTimestamp is the current timestamp from the dnsmasq logs.
	CurrentTimestamp int64

	// CurrentYear is the year this is happening.
	CurrentYear int

	// StatsCurrent is the current timestamp's stats.
	StatsCurrent *DNSStats

	// StatsPrevious is the last timestamp's stats.
	StatsPrevious *DNSStats
)

func init() {
	dnsmasqCmd.Flags().StringVarP(&DnsmasqLog, "log", "", dnsmasqLog, "dnsmasq log file.")
	dnsmasqCmd.Flags().BoolVarP(&FullLogs, "full", "", false, "Use full --log-queries logs.")
	RootCmd.AddCommand(dnsmasqCmd)
}

func dnsmasqSignalStats(t *tail.Tail, dog *statsd.Client) {
	// Set the current time from timestamp. Helps us to skip any items that are old.
	CurrentTimestamp = time.Now().Unix()
	StatsCurrent = new(DNSStats)
	StatsPrevious = new(DNSStats)

	go dnsmasqSignals()
	go setCurrentYear()
	for line := range t.Lines {
		// Blank lines really mess this up - this protects against it.
		if line.Text == "" {
			continue
		}
		// Parse line to grab timestamp - compare against CurrentTimestamp.
		// If it's older - skip. We would rather skip instead of double
		// count older data.
		if isOldTimestamp(line.Text) {
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
	// Check to see if we can send stats.
	// A new timestamp means we're getting new stats.
	checkStats()
	// Grab the timestamp from the log line.
	r := regexp.MustCompile(`\d+`)
	timestamp := r.FindString(content)
	unixTimestamp, _ := strconv.ParseInt(timestamp, 10, 64)
	CurrentTimestamp = unixTimestamp
	Log(fmt.Sprintf("StatsCurrent: %#v", StatsCurrent), "debug")
	StatsCurrent.timestamp = unixTimestamp
	Log(fmt.Sprintf("Timestamp: %d", unixTimestamp), "debug")
}

func checkStats() {
	// If we have actual stats in both Current and Previous.
	if (StatsCurrent.timestamp > 0) && (StatsPrevious.timestamp > 0) {
		// Let's send the stats to Datadog.
		SendSignalStats(*StatsCurrent, *StatsPrevious)
		Log(fmt.Sprintf("Current : %#v", StatsCurrent), "debug")
		Log(fmt.Sprintf("Previous: %#v", StatsPrevious), "debug")
		// Copy Current to Previous and zero out current.
		StatsPrevious = StatsCurrent
		StatsCurrent = new(DNSStats)
	} else if (StatsCurrent.timestamp > 0) && (StatsPrevious.timestamp == 0) {
		// We don't have enough stats to send.
		// Copy Current to Previous and zero out current.
		Log("Not enough stats to send.", "info")
		StatsPrevious = StatsCurrent
		StatsCurrent = new(DNSStats)
	} else if (StatsCurrent.timestamp == 0) && (StatsPrevious.timestamp == 0) {
		Log("Just starting up - nothing to do.", "info")
	}
}

// SendSignalStats sends stats to datadog using copies of the current data.
func SendSignalStats(current DNSStats, previous DNSStats) {
	Log("Sending stats now.", "info")
	Log(fmt.Sprintf("Current Copy : %#v", current), "debug")
	Log(fmt.Sprintf("Previous Copy: %#v", previous), "debug")
	forwards := current.queriesForwarded - previous.queriesForwarded
	locallyAnswered := current.queriesLocal - previous.queriesLocal
	dog := DogConnect()
	sendQueriesStats("dnsmasq.queries", forwards, "query:forward", dog)
	sendQueriesStats("dnsmasq.queries", locallyAnswered, "query:host", dog)
}

func sendQueriesStats(metric string, value int64, additionalTag string, dog *statsd.Client) {
	tags := dog.Tags
	dog.Tags = append(dog.Tags, additionalTag)
	dog.Count(metric, value, tags, signalInterval)
	dog.Tags = tags
}

func serverStats(content string) {
	r := regexp.MustCompile(`server (\d+\.\d+\.\d+\.\d+#\d+): queries sent (\d+), retried or failed (\d+)`)
	server := r.FindAllStringSubmatch(content, -1)
	if server != nil {
		srvr := server[0]
		serverAddress := srvr[1]
		serverAddressSent, _ := strconv.ParseInt(srvr[2], 10, 64)
		serverAddressRetryFailures, _ := strconv.ParseInt(srvr[3], 10, 64)
		serverStruct := DNSServer{timestamp: CurrentTimestamp, address: serverAddress, queriesSent: serverAddressSent, queriesFailed: serverAddressRetryFailures}
		StatsCurrent.servers = append(StatsCurrent.servers, serverStruct)
		Log(fmt.Sprintf("Time: %d Server: %s Queries: %d Retries/Failures: %d\n", CurrentTimestamp, serverAddress, serverAddressSent, serverAddressRetryFailures), "debug")
	}
}

func queriesForwarded(content string) {
	r := regexp.MustCompile(`forwarded (\d+),`)
	forwarded := r.FindAllStringSubmatch(content, -1)
	if forwarded != nil {
		fwd := forwarded[0]
		queriesForwarded, _ := strconv.ParseInt(fwd[1], 10, 64)
		StatsCurrent.queriesForwarded = queriesForwarded
		Log(fmt.Sprintf("Forwarded Queries: %d", queriesForwarded), "debug")
	}
}

func queriesLocal(content string) {
	r := regexp.MustCompile(`queries answered locally (\d+)`)
	local := r.FindAllStringSubmatch(content, -1)
	if local != nil {
		lcl := local[0]
		localResponses, _ := strconv.ParseInt(lcl[1], 10, 64)
		StatsCurrent.queriesLocal = localResponses
		Log(fmt.Sprintf("Responded Locally: %d", localResponses), "debug")
	}
}

func queriesAuthoritativeZones(content string) {
	r := regexp.MustCompile(`for authoritative zones (\d+)`)
	zones := r.FindAllStringSubmatch(content, -1)
	if zones != nil {
		zone := zones[0]
		authoritativeZones, _ := strconv.ParseInt(zone[1], 10, 64)
		StatsCurrent.authoritativeZones = authoritativeZones
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

func sendUSR1(procs []ProcessList) {
	if len(procs) > 0 {
		for _, proc := range procs {
			proc.USR1()
		}
	}
}

func getCurrentYear() int {
	t := time.Now()
	year := t.Year()
	Log(fmt.Sprintf("Year: %d", year), "debug")
	return year
}

func setCurrentYear() {
	for {
		CurrentYear = getCurrentYear()
		time.Sleep(time.Duration(yearSetInterval) * time.Second)
	}
}

func isOldTimestamp(line string) bool {
	// Munge the Syslog timestamp and pull out the values.
	dateTime := strings.TrimSpace(strings.Split(line, " dnsmasq")[0])
	dateTime = fmt.Sprintf("%s %d", dateTime, CurrentYear)
	stamp, _ := time.Parse("Jan _2 15:04:05 2006", dateTime)
	// If it's older than now - then skip it.
	if stamp.Unix() < CurrentTimestamp {
		Log(fmt.Sprintf("Skipping: '%s'", dateTime), "info")
		return true
	}
	return false
}
