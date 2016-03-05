// +build linux darwin freebsd

package cmd

import (
	"fmt"
	"github.com/DataDog/datadog-go/statsd"
	"github.com/hpcloud/tail"
	"github.com/spf13/cobra"
	"os"
	"regexp"
)

var tailCmd = &cobra.Command{
	Use:   "tail",
	Short: "Tail logs, match lines and send metrics to Datadog.",
	Long:  `Tail logs, match lines and send metrics to Datadog.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		checkTailFlags()
	},
	Run: startTail,
}

func startTail(cmd *cobra.Command, args []string) {
	// Try to compile the regex - throw an error if it doesn't work.
	regex, err := regexp.Compile(Match)
	if err != nil {
		fmt.Println("There's something wrong with your regex. Try again.")
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	dog := DogConnect()
	t := OpenLogfile(LogFile)
	TailLog(t, dog, regex)
}

func checkTailFlags() {
	if LogFile == "" {
		fmt.Println("Please enter a filename to tail '--log'")
		os.Exit(1)
	}
	if Match == "" {
		fmt.Println("Please enter a regex to match '--match'")
		os.Exit(1)
	}
	if MetricName == "" {
		fmt.Println("Please enter a metric name to send '--metric'")
		os.Exit(1)
	}
	fmt.Println("Press CTRL-C to shutdown.")
}

var (
	// LogFile is the file to tail.
	LogFile string

	// Match is the regex to match in the file.
	Match string

	// MetricName is the name of the metric to send to Datadog.
	MetricName string
)

func init() {
	tailCmd.Flags().StringVarP(&LogFile, "log", "", "", "File to tail.")
	tailCmd.Flags().StringVarP(&Match, "match", "", "", "Match this regex.")
	tailCmd.Flags().StringVarP(&MetricName, "metric", "", "", "Send this metric name.")
	RootCmd.AddCommand(tailCmd)
}

// TailLog tails a file and sends stats to Datadog.
func TailLog(t *tail.Tail, dog *statsd.Client, r *regexp.Regexp) {
	for line := range t.Lines {
		// Blank lines really mess this up - this protects against it.
		if line.Text == "" {
			continue
		}
		match := r.FindAllStringSubmatch(line.Text, -1)
		if match != nil {
			Log(fmt.Sprintf("Match: %s", match), "debug")
			Log(fmt.Sprintf("Sending Stat: %s", MetricName), "debug")
			dog.Count(MetricName, 1, dog.Tags, 1)
		}
	}
}
