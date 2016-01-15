// +build linux darwin freebsd

package cmd

import (
	"fmt"
	"github.com/hpcloud/tail"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/user"
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

// GetCurrentUsername grabs the current user running the binary.
func GetCurrentUsername() string {
	usr, _ := user.Current()
	username := usr.Username
	Log(fmt.Sprintf("username='%s'", username), "debug")
	return username
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

// OpenLogfile opens a logfile and passes back a *tail.Tail pointer.
func OpenLogfile(logfile string) *tail.Tail {
	t, err := tail.TailFile(logfile, tail.Config{
		ReOpen: true,
		Follow: true})
	if err != nil {
		Log("There was an error opening the file.", "info")
	}
	return t
}
