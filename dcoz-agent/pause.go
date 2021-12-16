package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

// Run pause program with given arguments
func pause(config *pauseConfig, duration time.Duration, targetPids []int) {
	pauseCmdRaw := []byte(fmt.Sprintf("%s %d %d %d %d",
		config.pauseBinPath, config.nCores, duration.Microseconds(), config.prio, len(targetPids)))
	for _, pid := range targetPids {
		pauseCmdRaw = append(pauseCmdRaw, []byte(fmt.Sprintf(" %d", pid))...)
	}
	pauseCmd := string(pauseCmdRaw)
	cmd := exec.Command("/bin/sh", "-c", pauseCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
