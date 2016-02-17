// +build linux darwin freebsd

package cmd

import (
	"fmt"
	"github.com/DataDog/datadog-go/statsd"
	"github.com/hpcloud/tail"
	"strings"
)

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

// SendLineStats sends the stats to Datadog.
func SendLineStats(dog *statsd.Client, line string, metric string) {
	Log(fmt.Sprintf("%s: %s", metric, line), "debug")
	oldTags := dog.Tags
	dog.Tags = append(dog.Tags, fmt.Sprintf("record:%s", metric))
	dog.Count("dnsmasq.event", 1, dog.Tags, 1)
	dog.Tags = oldTags
}
