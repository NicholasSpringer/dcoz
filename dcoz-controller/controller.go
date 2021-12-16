package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/NicholasSpringer/dcoz/dcoz-controller/experiment"
	"github.com/NicholasSpringer/dcoz/dcoz-controller/tracker"
	"github.com/NicholasSpringer/dcoz/dcoz-controller/util"
)

func main() {
	var pauseDuration time.Duration
	flag.DurationVar(&pauseDuration, "d", -1, "Duration of pause")

	var pausePeriod time.Duration
	flag.DurationVar(&pausePeriod, "p", -1, "Period of pause")

	var entrypoint string
	flag.StringVar(&entrypoint, "entrypoint", "", "IP of app entrypoint")

	flag.Parse()
	targets := flag.Args()

	errors := []string{}
	if pausePeriod == -1 {
		errors = append(errors, "Must specify -p")
	}
	if pauseDuration == -1 {
		errors = append(errors, "Must specify -d")
	}
	if entrypoint == "" {
		errors = append(errors, "Must specify -entrypoint")
	}
	if len(targets) == 0 {
		errors = append(errors, "Must specify list of targets as positional args")
	}
	if len(errors) != 0 {
		fmt.Println(strings.Join(errors, "\n"))
		os.Exit(1)
	}
	t := tracker.CreateTracker(entrypoint)
	expCfg := experiment.ExperimenterConfig{
		PauseDuration: pauseDuration,
		PausePeriod:   pausePeriod,
	}
	exp := experiment.CreateExperimenter(expCfg, t)
	stats, err := exp.RunExperiments(targets)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running experiment: %s\n", err)
		return
	}
	fmt.Println(stats)
	adjustedAvgLatencies := []float64{}
	for _, experimentStats := range stats {
		if experimentStats.NRequests == 0 {
			adjustedAvgLatencies = append(adjustedAvgLatencies, -1)
			continue
		}
		adjustedTotalLatency := util.AdjustLatency(experimentStats.TotalLatency, pauseDuration, pausePeriod)
		adjustedAvgLatency := (float64(adjustedTotalLatency) / float64(time.Millisecond)) / float64(experimentStats.NRequests)
		adjustedAvgLatencies = append(adjustedAvgLatencies, adjustedAvgLatency)
	}
	fmt.Println("Average Request Latencies:")
	for i := 0; i < len(targets); i += 1 {
		fmt.Printf("%s: %.3f ms\n", targets[i], adjustedAvgLatencies[i])
	}
	for {
	}
}
