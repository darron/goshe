// +build linux darwin freebsd

package cmd

import (
	"fmt"
	"github.com/darron/datadog-go/statsd"
	"github.com/spf13/cobra"
	"github.com/tatsushid/go-fastping"
	"net"
	"os"
	"strings"
	"time"
)

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Ping an address and send stats to Datadog.",
	Long:  `Ping an address and send stats to Datadog. Need to be root to use.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		checkPingFlags()
	},
	Run: startPing,
}

func startPing(cmd *cobra.Command, args []string) {
	dog := DogConnect()
	for {
		go ping(Endpoint, dog)
		time.Sleep(time.Duration(Interval) * time.Second)
	}
}

func checkPingFlags() {
	if Endpoint == "" {
		fmt.Println("Please enter an address or domain name to ping: -e")
		os.Exit(1)
	}
	if !checkRootUser() {
		fmt.Println("You need to be root to run this.")
		os.Exit(1)
	}
	fmt.Println("Press CTRL-C to shutdown.")
}

var (
	// Endpoint holds the address we're going to ping.
	Endpoint string
)

func init() {
	pingCmd.Flags().StringVarP(&Endpoint, "endpoint", "e", "www.google.com", "Endpoint to ping.")
	RootCmd.AddCommand(pingCmd)
}

func checkRootUser() bool {
	user := GetCurrentUsername()
	if user != "root" {
		return false
	}
	return true
}

func ping(address string, dog *statsd.Client) {
	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", address)
	if err != nil {
		fmt.Println(err)
	}
	p.AddIPAddr(ra)
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		fmt.Printf("IP Addr: %s receive, RTT: %v\n", addr.String(), rtt)
		go sendPingStats(dog, rtt)
	}
	err = p.Run()
	if err != nil {
		fmt.Println(err)
	}
}

func sendPingStats(dog *statsd.Client, rtt time.Duration) {
	var err error
	seconds := (float64(rtt) / 1000000000)
	address := strings.ToLower(strings.Replace(Endpoint, ".", "_", -1))
	metricName := fmt.Sprintf("ping.%s", address)
	err = dog.Histogram(metricName, seconds, dog.Tags, 1)
	if err != nil {
		Log(fmt.Sprintf("Error sending ping stats for '%s'", Endpoint), "info")
	}
}
