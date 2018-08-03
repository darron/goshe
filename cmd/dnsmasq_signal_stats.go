// +build linux darwin freebsd

package cmd

import (
	"fmt"
	"github.com/DataDog/datadog-go/statsd"
	"github.com/hpcloud/tail"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
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

// dnsmasqSignalStats processes the logs that are output by dnsmasq
// when the USR1 signal is sent to it.
func dnsmasqSignalStats(t *tail.Tail) {
	// Set the current time from timestamp. Helps us to skip any items that are old.
	CurrentTimestamp = time.Now().Unix()
	StatsCurrent = new(DNSStats)
	StatsPrevious = new(DNSStats)

	go dnsmasqSignals()
	if MemoryFlag {
		go sendDnsmasqMemStats()
	}
	for line := range t.Lines {
		// Blank lines really mess this up - this protects against it.
		if line.Text == "" {
			continue
		}
		// Let's process the lines.
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
		if strings.HasPrefix(content, "read") {
			Log(fmt.Sprintf("line: %s", content), "debug")
			readStats(content)
		}
	}
}

// grabTimestamp pulls the timestamp out of the logs and checks
// to see if we can send stats via checkStats()/
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

// checkStats looks to see if we have current and previous stats and
// then does what's appropriate.
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

// SendSignalStats sends stats to Datadog using copies of the current data.
// TODO: Right now we're ignoring all sorts of stats - will see if we need them.
func SendSignalStats(current DNSStats, previous DNSStats) {
	Log("Sending stats now.", "debug")
	Log(fmt.Sprintf("Current Copy : %#v", current), "debug")
	Log(fmt.Sprintf("Previous Copy: %#v", previous), "debug")
	forwards := current.queriesForwarded - previous.queriesForwarded
	locallyAnswered := current.queriesLocal - previous.queriesLocal
	dog := DogConnect()
	// Make sure the stats are positive - if they're negative dnsmasq must have been
	// restarted and those numbers will not be accurate.
	if forwards >= 0 {
		sendQueriesStats("dnsmasq.queries", forwards, "query:forward", dog)
		Log(fmt.Sprintf("Forwards: %d", forwards), "debug")
	} else {
		Log("Negative forwarded queries detected - dnsmasq must have been restarted.", "info")
		sendQueriesStats("dnsmasq.queries", current.queriesForwarded, "query:forward", dog)
		Log(fmt.Sprintf("Forwards: %d", current.queriesForwarded), "debug")
	}
	if locallyAnswered >= 0 {
		sendQueriesStats("dnsmasq.queries", locallyAnswered, "query:local", dog)
		Log(fmt.Sprintf("Locally Answered: %d", locallyAnswered), "debug")
	} else {
		Log("Negative locally answered queries detected - dnsmasq must have been restarted.", "info")
		sendQueriesStats("dnsmasq.queries", current.queriesLocal, "query:local", dog)
		Log(fmt.Sprintf("Locally Answered: %d", current.queriesLocal), "debug")
	}
}

// sendQueriesStats actually sends the stats to Dogstatsd.
func sendQueriesStats(metric string, value int64, additionalTag string, dog *statsd.Client) {
	tags := dog.Tags
	dog.Tags = append(dog.Tags, additionalTag)
	if os.Getenv("GOSHE_ADDITIONAL_TAGS") != "" {
		dog.Tags = append(dog.Tags, os.Getenv("GOSHE_ADDITIONAL_TAGS"))
	}
	dog.Count(metric, value, tags, signalInterval)
	dog.Tags = tags
}

// serverStats gets the stats for a DNSServer struct.
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

// queriesForwarded gets how many queries are forwarded to a DNSServer
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

// queriesLocal gets how many queries are answered locally. Hosts files
// are included.
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

// queriesAuthoritativeZones gets how many authoritative zones are present.
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

func sendHistogramStats(metric string, value float64, additionalTag string, dog *statsd.Client) {
	tags := dog.Tags
	dog.Tags = append(dog.Tags, additionalTag)
	if os.Getenv("GOSHE_ADDITIONAL_TAGS") != "" {
		dog.Tags = append(dog.Tags, os.Getenv("GOSHE_ADDITIONAL_TAGS"))
	}
	dog.Histogram(metric, value, tags, 1)
	dog.Tags = tags
}

func readStats(content string) {
	filename, addresses := LoadFilesStats(content)
	if filename != "" && addresses > 0 {
		Log("Sending the loaded file stats", "debug")
		dog := DogConnect()
		fileTag := fmt.Sprintf("filename:%s", filename)
		sendHistogramStats("dnsmasq.hosts_file_stats", addresses, fileTag, dog)
	}
}

// LoadFilesStats - a testable function to get the loaded file stats.
func LoadFilesStats(content string) (string, float64) {
	r := regexp.MustCompile(`read (.*) - (\d+) addresses`)
	domainsLoaded := r.FindAllStringSubmatch(content, -1)
	if domainsLoaded != nil {
		pieces := domainsLoaded[0]
		filename := pieces[1]
		addresses, _ := strconv.ParseFloat(pieces[2], 64)
		Log(fmt.Sprintf("Filename: %s, Addresses: %f", filename, addresses), "debug")
		return filename, addresses
	}
	return "", 0
}

// dnsmasqSignals loops and send USR1 to each dnsmasq process
// after each signalInterval - USR1 outputs logs with statistics.
func dnsmasqSignals() {
	for {
		procs := GetMatches("/usr/sbin/dnsmasq", false)
		// If we've defined this ENV VAR - then we do NOT want to send
		// signals. It's a way to run multiple versions at the same time.
		if os.Getenv("GOSHE_DISABLE_DNSMASQ_SIGNALS") == "" {
			sendUSR1(procs)
		}
		time.Sleep(time.Duration(signalInterval) * time.Second)
	}
}

// sendUSR1 actually sends the signal.
func sendUSR1(procs []ProcessList) {
	if len(procs) > 0 {
		for _, proc := range procs {
			proc.USR1()
		}
	}
}

// sendDnsmasqMemStats sends memory stats for dnsmasq if MemoryFlag
// is true.
func sendDnsmasqMemStats() {
	for {
		matches := GetMatches("/usr/sbin/dnsmasq", false)
		if matches != nil {
			fmt.Printf("Found %d matches.\n", len(matches))
			SendMetrics(matches)
		} else {
			fmt.Printf("Did not find any matches.\n")
		}
		time.Sleep(time.Duration(Interval) * time.Second)
	}
}
