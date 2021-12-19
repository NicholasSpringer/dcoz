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
	fmt.Println("Sleeping 10s")
	time.Sleep(time.Second * 10)
	var pauseDuration time.Duration
	flag.DurationVar(&pauseDuration, "d", -1, "Duration of pause")

	var pausePeriod time.Duration
	flag.DurationVar(&pausePeriod, "p", -1, "Period of pause")

	var requestPeriod time.Duration
	flag.DurationVar(&requestPeriod, "rp", -1, "Period of requests")

	var expDuration time.Duration
	flag.DurationVar(&expDuration, "ed", -1, "Duration of each experiment")

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
	t := tracker.CreateTracker(entrypoint, requestPeriod)
	expCfg := experiment.ExperimenterConfig{
		PauseDuration: pauseDuration,
		PausePeriod:   pausePeriod,
		ExpDuration:   expDuration,
	}
	exps := []experiment.ExperimenterConfig{expCfg}
	ids := []string{"1"}
	for i := 0; i < len(exps); i++ {
		runExperimenter(ids[i], targets, t, exps[i])
	}
	t.StartTracking()
	time.Sleep(time.Second * 120)
	stats := t.FinishTracking()
	avgLatency := (float64(stats.TotalLatency) / float64(time.Millisecond)) / float64(stats.NRequests)
	fmt.Println()
	fmt.Println("No pauses")
	fmt.Printf("%.3f ms\n", avgLatency)
	t.StopWorkload()
	for {
	}
}

func runExperimenter(expId string, targets []string, t *tracker.Tracker, expCfg experiment.ExperimenterConfig) {
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
		adjustedTotalLatency := util.AdjustLatency(experimentStats.TotalLatency, expCfg.PauseDuration, expCfg.PausePeriod)
		adjustedAvgLatency := (float64(adjustedTotalLatency) / float64(time.Millisecond)) / float64(experimentStats.NRequests)
		adjustedAvgLatencies = append(adjustedAvgLatencies, adjustedAvgLatency)
	}
	fmt.Println()
	fmt.Printf("EXP (%s) Average Request Latencies:", expId)
	for i := 0; i < len(targets); i += 1 {
		fmt.Printf("%s: %.3f ms\n", targets[i], adjustedAvgLatencies[i])
	}
	fmt.Println()
	return
}
