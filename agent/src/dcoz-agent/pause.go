package main

import (
	"fmt"
	"os"
	"os/exec"
)

// Run pause program with given arguments
func pause(config *pauseConfig, duration int, targetPids []int) {
	pauseCmdRaw := []byte(fmt.Sprintf("%s %d %d %d %d",
		config.pauseBinPath, config.nCores, duration, config.prio, len(targetPids)))
	for _, pid := range targetPids {
		pauseCmdRaw = append(pauseCmdRaw, []byte(fmt.Sprintf(" %d", pid))...)
	}
	pauseCmd := string(pauseCmdRaw)
	println(pauseCmd)
	cmd := exec.Command("/bin/sh", "-c", pauseCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("pause error: ", err.Error())
		os.Exit(1)
	}
}
