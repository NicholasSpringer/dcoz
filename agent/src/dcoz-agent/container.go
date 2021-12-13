package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"time"
)

// Translate target pod identifiers to a list of associated pids
func getTargetPids(targetContainers []string, procPath string) []int {
	start := time.Now()
	// Create set from target pod list
	targetContainersSet := make(map[string]struct{})
	for _, pod := range targetContainers {
		targetContainersSet[pod] = struct{}{}
	}
	t1 := time.Since(start)
	procDirs, err := ioutil.ReadDir(procPath)
	if err != nil {
		fmt.Println("error reading proc dir: ", err.Error())
		os.Exit(1)
	}
	t2 := time.Since(start)
	targetPids := []int{}
	for _, procDir := range procDirs {
		pid, err := strconv.Atoi(procDir.Name())
		if err != nil {
			// Directory not a process id
			continue
		}
		// Determine container identifier of pid
		containerId := getContainerId(procPath, pid)
		if containerId == "" {
			// Process may no longer exist or may not be in a docker container
			continue
		}
		// Check if container id is in targets
		_, isTarget := targetContainersSet[containerId]
		fmt.Printf("%s %t\n", containerId, isTarget)
		if isTarget {
			targetPids = append(targetPids, pid)
		}
	}
	t3 := time.Since(start)
	fmt.Println(targetPids)
	fmt.Println(t1.Milliseconds(), " ", t2.Milliseconds(), " ", t3.Milliseconds())
	return targetPids
}

var cgroupPattern *regexp.Regexp = regexp.MustCompile(
	`\d+:[^:]+:\/kubepods\/[^/]+\/[^/]+\/([a-z|\d]{64})`)

func getContainerId(procPath string, pid int) string {
	cgroupFile, err := os.Open(fmt.Sprintf("%s/%d/cgroup", procPath, pid))
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
