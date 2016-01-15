// +build linux darwin freebsd

package cmd

import (
	"fmt"
	"github.com/DataDog/datadog-go/statsd"
	"github.com/spf13/cobra"
	"strings"
)

const (
	dnsmasqLog = "/var/log/dnsmasq/dnsmasq"
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

func checkDnsmasqFlags() {
	fmt.Println("Press CTRL-C to shutdown.")
}

var (
	// DnsmasqLog is the logfile that dnsmasq logs to.
	DnsmasqLog string
)

func init() {
	dnsmasqCmd.Flags().StringVarP(&DnsmasqLog, "log", "", dnsmasqLog, "dnsmasq log file.")
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
