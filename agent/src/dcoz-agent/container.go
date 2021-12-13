package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
)

// Translate target pod identifiers to a list of associated pids
func getTargetPids(targetContainers []string) []int {
	// Create set from target pod list
	targetContainersSet := make(map[string]struct{})
	for _, pod := range targetContainers {
		targetContainersSet[pod] = struct{}{}
	}
	procDirs, err := ioutil.ReadDir("/proc")
	if err != nil {
		fmt.Println("error reading proc dir: ", err.Error())
		os.Exit(1)
	}
	targetPids := []int{}
	for _, procDir := range procDirs {
		pid, err := strconv.Atoi(procDir.Name())
		if err != nil {
			// Directory not a process id
			continue
		}
		// Determine container identifier of pid
		containerId := getContainerId(pid)
		if containerId == "" {
			// Process may no longer exist or may not be in a docker container
			continue
		}
		// Check if container id is in targets
		_, isTarget := targetContainersSet[containerId]
		if isTarget {
			targetPids = append(targetPids, pid)
		}
	}
	return targetPids
}

var cgroupPattern *regexp.Regexp = regexp.MustCompile(
	`\d+:[^:]+:\/kubepods\/[^/]+\/[^/]+\/([a-z|\d]{64})`)

func getContainerId(pid int) string {
	cgroupFile, err := os.Open(fmt.Sprintf("/proc/%d/cgroup", pid))
	if err != nil {
		// Pid may no longer exist
		fmt.Println("error reading cgroup file: ", err.Error())
		return ""
	}
	scanner := bufio.NewScanner(cgroupFile)
	for scanner.Scan() {
		matches := cgroupPattern.FindStringSubmatch(scanner.Text())
		if matches != nil {
			return matches[1]
		}
	}
	// Return empty string if no container id found
	return ""
}
