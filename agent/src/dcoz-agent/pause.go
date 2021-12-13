package main

import (
	"fmt"
	"os"
	"os/exec"
)

// Run pause program with given arguments
func pause(procPath string, config *pauseConfig, duration int, targetPids []int) {
	targetNsProcPath := fmt.Sprintf("%s/%d", procPath, PAUSE_TARGET_NS_PID)
	pauseCmdRaw := []byte(fmt.Sprintf("%s %s %d %d %d %d",
		config.pauseBinPath, targetNsProcPath, config.nCores, duration, config.prio, len(targetPids)))
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
