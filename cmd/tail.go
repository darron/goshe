// +build linux darwin freebsd

package cmd

import (
	"bufio"
	"fmt"
	"github.com/DataDog/datadog-go/statsd"
	"github.com/hpcloud/tail"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var tailCmd = &cobra.Command{
	Use:   "tail",
	Short: "Tail logs or stdout, match lines and send metrics to Datadog.",
	Long:  `Tail logs or stdout, match lines and send metrics to Datadog.`,
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
	// For the Logfile option.
	if LogFile != "" {
		t := OpenLogfile(LogFile)
		TailLog(t, dog, regex)
	}
	// If you're capturing stdout of a program.
	if ProgramStdout != "" {
		TailOutput(dog, regex)
	}
}

func checkTailFlags() {
	if LogFile == "" && ProgramStdout == "" {
		fmt.Println("Please enter a filename to tail '--log' OR a program to run '--program'")
		os.Exit(1)
	}
	if LogFile != "" && ProgramStdout != "" {
		fmt.Println("Please choose '--log' OR '--program' - you cannot have both.")
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
	// If you're sending a MetricTag - it needs to have a ':'
	if MetricTag != "" && !strings.Contains(MetricTag, ":") {
		fmt.Println("Tags need to contain a ':'")
		os.Exit(1)
	}
	fmt.Println("Press CTRL-C to shutdown.")
}

var (
	// LogFile is the file to tail.
	LogFile string

	// ProgramStdout is a program to run to capture stdout.
	ProgramStdout string

	// Match is the regex to match in the file.
	Match string

	// MetricName is the name of the metric to send to Datadog.
	MetricName string

	// MetricTag is the name of the tag to add to the metric we're sending to Datadog.
	MetricTag string
)

func init() {
	tailCmd.Flags().StringVarP(&LogFile, "log", "", "", "File to tail.")
	tailCmd.Flags().StringVarP(&ProgramStdout, "program", "", "", "Program to run for stdout.")
	tailCmd.Flags().StringVarP(&Match, "match", "", "", "Match this regex.")
	tailCmd.Flags().StringVarP(&MetricName, "metric", "", "", "Send this metric name.")
	tailCmd.Flags().StringVarP(&MetricTag, "tag", "", "", "Add this tag to the metric.")
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
			tags := dog.Tags
			if MetricTag != "" {
				tags = append(tags, MetricTag)
			}
			dog.Count(MetricName, 1, tags, 1)
		}
	}
}

// TailOutput watches the output of ProgramStdout and matches on those lines.
func TailOutput(dog *statsd.Client, r *regexp.Regexp) {
	cli, args := processCommand(ProgramStdout)
	runCommand(cli, args, r, dog)
}

func processCommand(command string) (string, []string) {
	var cli string
	var args []string

	parts := strings.Fields(command)
	cli = parts[0]
	args = parts[1:]
	Log(fmt.Sprintf("Cli: %s Args: %s", cli, args), "debug")

	return cli, args
}

func runCommand(cli string, args []string, r *regexp.Regexp, dog *statsd.Client) {
	cmd := exec.Command(cli, args...)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		Log(fmt.Sprintf("There was an error running '%s': %s", ProgramStdout, err), "info")
		os.Exit(1)
	}
	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			Log(fmt.Sprintf("Line: %s", line), "debug")
			// Blank lines are bad for the matching software - it freaks out.
			if line == "" {
				continue
			}
			match := r.FindAllStringSubmatch(line, -1)
			if match != nil {
				Log(fmt.Sprintf("Match: %s", match), "debug")
				Log(fmt.Sprintf("Sending Stat: %s", MetricName), "debug")
				tags := dog.Tags
				if MetricTag != "" {
					tags = append(tags, MetricTag)
				}
				dog.Count(MetricName, 1, tags, 1)
			}
		}
	}()

	err = cmd.Start()
	if err != nil {
		Log("There was and error starting the command.", "info")
	}

	err = cmd.Wait()
	if err != nil {
		Log("There was and error waiting for the command.", "info")
	}
}
