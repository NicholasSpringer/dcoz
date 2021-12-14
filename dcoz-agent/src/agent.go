package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

type pauseConfig struct {
	nCores       int
	prio         int
	pauseBinPath string
}

type agent struct {
	pauseCfg pauseConfig
	port     int
}

type speedupMessage struct {
	Duration         int      `json:"duration"`
	TargetContainers []string `json:"targetContainers"`
}

const PAUSE_TARGET_NS_PID = 1

func main() {
	var nCores int
	flag.IntVar(&nCores, "cores", 0, "Number of cores in the system")

	var prio int
	flag.IntVar(&prio, "priority", -1, "Priority of blocker threads")

	var pauseBinPath string
	flag.StringVar(&pauseBinPath, "pause", "", "Pausing binary path")

	var port int
	flag.IntVar(&port, "port", -1, "Port to listen on")

	flag.Parse()
	errors := []string{}
	if nCores == 0 {
		errors = append(errors, "Must specify num cores using -cores")
	}
	if prio == -1 {
		errors = append(errors, "Must specify blocker thread priority using -priority")
	}
	if pauseBinPath == "" {
		errors = append(errors, "Must specify pause path using -pause")
	}
	if port == -1 {
		errors = append(errors, "Must specify port using -port")
	}
	if len(errors) != 0 {
		fmt.Println(strings.Join(errors, "\n"))
		os.Exit(1)
	}
	ag := agent{pauseCfg: pauseConfig{nCores: nCores,
		prio: prio, pauseBinPath: pauseBinPath}, port: port}
	ag.listen()
}

// Listen for messages from controller
func (ag *agent) listen() {
	addr := net.UDPAddr{
		Port: ag.port,
		IP:   net.ParseIP("0.0.0.0"),
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer conn.Close()
	buff := make([]byte, 2048)
	println("Agent started!")
	for {
		length, contrAddr, err := conn.ReadFromUDP(buff)
		if err != nil {
			fmt.Println("udp listening error: ", err.Error())
			os.Exit(1)
		}
		var speedupMsg speedupMessage
		err = json.Unmarshal(buff[:length], &speedupMsg)
		if err != nil {
			fmt.Println("json unmarshal error: ", err)
			continue
		}
		fmt.Printf("Speedup Message: %v\n", speedupMsg)
		// Respond with heartbeat after successfully receiving message
		hbMsg := []byte{0}
		conn.WriteToUDP(hbMsg, contrAddr)

		// If pause duration is 0, do not execute a pause. Otherwise, translate
		// target pod identifier to target local pids and execute pause.
		if speedupMsg.Duration != 0 {
			targetPids := getTargetPids(speedupMsg.TargetContainers)
			pause(&ag.pauseCfg, speedupMsg.Duration, targetPids)
		}
	}
}
