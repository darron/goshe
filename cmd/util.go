// +build linux darwin freebsd

package cmd

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
	"strings"
	"time"
)

// ReturnCurrentUTC returns the current UTC time in RFC3339 format.
func ReturnCurrentUTC() string {
	t := time.Now().UTC()
	dateUpdated := (t.Format(time.RFC3339))
	return dateUpdated
}

// SetDirection returns the direction.
func SetDirection() string {
	args := fmt.Sprintf("%x", os.Args)
	direction := "main"
	if strings.ContainsAny(args, " ") {
		if strings.HasPrefix(os.Args[1], "-") {
			direction = "main"
		} else {
			direction = os.Args[1]
		}
	}
	return direction
}

// Log adds the global Direction to a message and sends to syslog.
// Syslog is setup in main.go
func Log(message, priority string) {
	message = fmt.Sprintf("%s: %s", Direction, message)
	if Verbose {
		time := ReturnCurrentUTC()
		fmt.Printf("%s: %s\n", time, message)
	}
	switch {
	case priority == "debug":
		if os.Getenv("GOSHE_DEBUG") != "" {
			log.Print(message)
		}
	default:
		log.Print(message)
	}
}

// GetHostname returns the hostname.
func GetHostname() string {
	hostname, _ := os.Hostname()
	return hostname
}

// LoadConfig loads the configuration from a config file.
func LoadConfig() {
	Log("Loading viper config.", "info")
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/goshe/")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		Log(fmt.Sprintf("No config file found: %s \n", err), "info")
	}
	viper.SetEnvPrefix("GOSHE")
}
