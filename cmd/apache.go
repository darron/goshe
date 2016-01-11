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

func init() {
	RootCmd.AddCommand(apacheCmd)
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
		if pid != "-" && !strings.HasPrefix(vhost, "*") {
			pidInt, _ := strconv.ParseInt(pid, 10, 64)
			apache = ApacheProcess{Pid: pidInt, Vhost: vhost}
			stats = append(stats, apache)
		}
	}
	return stats
}
