// +build linux darwin freebsd

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	modStatus = "http://127.0.0.1/server-status/"
	minMemory = 10 * 1024 * 1024 // Skip logging Apache processes with less than this memory used: 10MB by default.
)

// ApacheProcess holds the interesting pieces of Apache stats.
type ApacheProcess struct {
	Pid   int64
	Vhost string
}

var apacheCmd = &cobra.Command{
	Use:   "apache",
	Short: "Grab stats from Apache2 processes - and mod_status - and sends to Datadog.",
	Long:  `Grab stats from Apache2 processes - and mod_status - and sends to Datadog.`,
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
			processMap := createProcessMemMap(matches)
			// Let's get the Apache details and then submit those.
			apaches := GetApacheServerStats()
			SendApacheServerStats(apaches, processMap)
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

var (
	// MinimumMemory is the minimum size of an Apache process that we log stats for.
	MinimumMemory uint64
)

func init() {
	apacheCmd.Flags().Uint64VarP(&MinimumMemory, "memory", "m", minMemory, "Smallest Apache memory size to log.")
	RootCmd.AddCommand(apacheCmd)
}

// SendApacheServerStats sends tagged Apache stats to Datadog
func SendApacheServerStats(apache []ApacheProcess, procs map[int]uint64) {
	var err error
	dog := DogConnect()
	for _, server := range apache {
		Log(fmt.Sprintf("SendApacheServerStats server='%#v'", server), "debug")
		pid := int(server.Pid)
		memory := procs[pid]
		if memory > MinimumMemory {
			Log(fmt.Sprintf("sending memory='%f' vhost='%s'", float64(memory), server.Vhost), "debug")
			dog.Tags = append(dog.Tags, fmt.Sprintf("site:%s", server.Vhost))
			err = dog.Histogram("apache2.rss_memory_tagged", float64(memory), dog.Tags, 1)
			if err != nil {
				Log("Error sending tagged rss_memory stats for Apache", "info")
			}
		}
	}
}

// GetApacheServerStats grabs info from mod_status and parses it.
func GetApacheServerStats() []ApacheProcess {
	htmlContent := getServerStatus("")
	stringResults := parseServerStatus(htmlContent)
	apaches := parseProcessStats(stringResults)
	return apaches
}

// getServerStatus returns the HTML nodes from the Apache serverStatus page.
func getServerStatus(server string) *html.Node {
	serverStatus := ""
	if serverStatus = viper.GetString("apache_status"); serverStatus == "" {
		serverStatus = modStatus
	}
	if server != "" {
		serverStatus = server
	}
	request, err := http.NewRequest("GET", serverStatus, nil)
	if err != nil {
		Log("Error connecting to the Apache server.", "info")
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		Log("Error reading the response.", "info")
	}
	root, err := html.Parse(response.Body)
	if err != nil {
		Log("Something went wrong with the html parse.", "info")
	}
	defer response.Body.Close()
	return root
}

// parseServerStatus returns a slice of strings containing only server stats.
func parseServerStatus(root *html.Node) []string {
	var apacheStats []string
	// Lines with stats start with a number.
	var validStats = regexp.MustCompile(`^[0-9]`)
	// Grab all the table rows.
	rows := scrape.FindAll(root, scrape.ByTag(atom.Tr))
	// If each row matches - add it to the stats lines.
	for _, row := range rows {
		content := scrape.Text(row)
		if validStats.MatchString(content) {
			apacheStats = append(apacheStats, content)
		}
	}
	Log(fmt.Sprintf("parseServerStatus apacheStats='%d'", len(apacheStats)), "debug")
	return apacheStats
}

// parseProcessStats takes the slice of strings and returns a slice of Apache processes.
func parseProcessStats(processes []string) []ApacheProcess {
	var stats []ApacheProcess
	var apache ApacheProcess
	for _, process := range processes {
		fields := strings.SplitAfterN(process, " ", -1)
		pid := strings.TrimSpace(fields[1])
		vhost := strings.TrimSpace(fields[11])
		if pid != "-" && !strings.HasPrefix(vhost, "*") && vhost != "?" {
			pidInt, _ := strconv.ParseInt(pid, 10, 64)
			apache = ApacheProcess{Pid: pidInt, Vhost: vhost}
			stats = append(stats, apache)
		}
	}
	Log(fmt.Sprintf("parseProcessStats stats='%d'", len(stats)), "debug")
	return stats
}

// Take a slice of ProcessList items and create a map.
func createProcessMemMap(p []ProcessList) map[int]uint64 {
	m := make(map[int]uint64)
	for _, proc := range p {
		pid := proc.Pid
		mem := proc.Pmem
		Log(fmt.Sprintf("pid='%d' mem='%d'", pid, mem), "debug")
		m[pid] = mem
	}
	return m
}
