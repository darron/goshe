// +build linux darwin freebsd

package cmd

import (
	"fmt"
	"github.com/DataDog/datadog-go/statsd"
)

const (
	// DogStatsdAddr is the default address for Dogstatsd.
	DogStatsdAddr = "127.0.0.1:8125"
)

// DogConnect sets up a connection and sets standard tags.
func DogConnect() *statsd.Client {
	connection := DogStatsdSetup()
	return connection
}

// DogStatsdSetup sets up a connection to DogStatsd.
func DogStatsdSetup() *statsd.Client {
	c, err := statsd.New(DogStatsdAddr)
	if err != nil {
		Log(fmt.Sprintf("DogStatsdSetup Error: %#v", err), "info")
	}
	c.Namespace = fmt.Sprintf("%s.", MetricPrefix)
	return c
}
