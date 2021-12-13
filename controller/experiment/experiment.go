package experiment

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/dcoz-controller/server"
	"go.uber.org/atomic"
)

// experiment contains information about a SINGLE virutal speed up,
// multiple experiments are run with different virutal speedups applied
type Experiment struct {
	numRequests        int
	successfulRequests atomic.Int64
	averageLatency     float64
	virtualSpeedUp     float64
	targetService      string
	targetServiceIP    *net.UDPAddr
	errChan            chan bool
	latencyChan        chan float64
	fields             []string
	workload           [][]string
	udpServer          *server.UDPServer
}

func CreateExperiment(server *server.UDPServer, numRequests int, virtualSpeedUp float64, targetService string, targetServiceIP *net.UDPAddr, workload [][]string) *Experiment {
	// buffered channel the size of the number of requests required to run experiment
	errChan := make(chan bool, numRequests)
	latencyChan := make(chan float64, numRequests)
	return &Experiment{
		numRequests:     numRequests,
		virtualSpeedUp:  virtualSpeedUp,
		fields:          workload[0],
		workload:        workload[1:],
		targetService:   targetService,
		targetServiceIP: targetServiceIP,
		errChan:         errChan,
		latencyChan:     latencyChan,
		udpServer:       server,
	}
}

func (e *Experiment) RunExperiment() {
	// apply virtual slowdown here
	e.udpServer.BroadcastSpeedMsg(e.virtualSpeedUp, e.targetServiceIP)

	for _, requestVals := range e.workload {
		form := createPostForm(e.fields, requestVals)
		go e.makeRequest(form)
	}
	// waiting for the correct number of requests
	for len(e.errChan)+len(e.latencyChan) < e.numRequests {

	}

	// revert
	e.udpServer.BroadcastSpeedMsg(1, e.targetServiceIP)
	// log results
}

func (e *Experiment) LogResults() {

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

func (e *Experiment) makeRequest(form []byte) {
	start := time.Now()
	_, err := http.Post("IP HERE", "application/json", bytes.NewBuffer(form))
	if err != nil {
		e.errChan <- true
		return
	}
	elapsed := time.Now().Sub(start)
	e.latencyChan <- float64(elapsed)
}
