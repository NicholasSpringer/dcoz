package experiment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dcoz-controller/server"
	"github.com/dcoz-controller/utils"
)

// experiment contains information about a SINGLE virutal speed up,
// multiple experiments are run with different virutal speedups applied
type Experiment struct {
	numRequests      int
	slowDownDuration int
	targetContainers []string
	entryPoint       string
	fields           []string
	workload         [][]string
	udpServer        *server.UDPServer
	requests         map[int][]float64
}

func createExperiment(server *server.UDPServer, targetContainers []string, entryPoint string, workload [][]string) *Experiment {
	// buffered channel the size of the number of requests required to run experiment
	requests := make(map[int][]float64)
	fields := workload[0]
	workloadValues := workload[1:]
	for i := 0; i < len(workloadValues); i++ {
		requests[i] = []float64{0, 0}
	}
	return &Experiment{
		numRequests:      len(workload),
		slowDownDuration: utils.SLOWDOWNDURATION,
		fields:           fields,
		workload:         workloadValues,
		targetContainers: targetContainers,
		entryPoint:       entryPoint,
		udpServer:        server,
		requests:         requests,
	}
}

func (e *Experiment) runExperiment() (float64, int) {
	reqCt := 0
	pauseTicker := time.NewTicker(utils.NVALUE * time.Millisecond)
	reqTicker := time.NewTicker(utils.KVALUE * time.Millisecond)
	done := make(chan bool, 1)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-pauseTicker.C:
				e.udpServer.BroadcastPause(float64(e.slowDownDuration), e.targetContainers)
				e.requests[reqCt][1] += utils.SLOWDOWNDURATION
			}
		}
	}()
	for reqCt < e.numRequests {
		<-reqTicker.C
		form := createPostForm(e.fields, e.workload[reqCt])
		latency, err := e.makeRequest(form, e.entryPoint)
		if err != nil {
			// if the request fails for any reason, discard request from count
			e.requests[reqCt][0] = -1
		} else {
			e.requests[reqCt][0] = latency
		}
		reqCt += 1
	}
	done <- true
	reqTicker.Stop()
	pauseTicker.Stop()

	// log results
	return e.getAvgVirtualLatency()
}

func getVirtualLatency(latency, totalDuration float64) float64 {
	return latency - totalDuration
}

func (e *Experiment) getAvgVirtualLatency() (float64, int) {
	reqs := 0
	avgLatency := float64(0)
	for _, values := range e.requests {
		latency := values[0]
		duration := values[1]
		if latency != -1 {
			virtualLatency := getVirtualLatency(latency, duration)
			avgLatency += virtualLatency
			reqs += 1
		}
	}
	return avgLatency / float64(reqs), reqs
}

func createPostForm(fields, values []string) []byte {
	l := len(fields)
	var form map[string]string
	for i := 0; i < l; i++ {
		form[fields[i]] = values[i]
	}
	ret, _ := json.Marshal(form)
	return ret
}

func (e *Experiment) makeRequest(form []byte, entryPoint string) (float64, error) {
	start := time.Now()
	_, err := http.Post(entryPoint, "application/json", bytes.NewBuffer(form))
	if err != nil {
		return -1, err
	}
	elapsed := time.Now().Sub(start)
	return float64(elapsed), nil
}

func RunExperiments(server *server.UDPServer, workload [][]string, containers []string, entryPoint string) error {
	// creating output file
	file, err := os.Create("dcoz_results.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	// sequentially running
	for _, container := range containers {
		targetContainers := []string{container}
		e := createExperiment(server, targetContainers, entryPoint, workload)
		avgVirtualLatency, reqs := e.runExperiment()

		tmpString := fmt.Sprintf("Target Container(s): %s\n", strings.Join(targetContainers, ","))
		file.WriteString(tmpString)
		tmpString = fmt.Sprintf("Average Virtual Latency: %.2f\n", avgVirtualLatency)
		file.WriteString(tmpString)
		tmpString = fmt.Sprintf("Requests: %d\n", reqs)
		file.WriteString(tmpString)
		file.WriteString("\n\n")
	}
	return nil
}
