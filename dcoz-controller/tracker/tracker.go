package tracker

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/atomic"
)

const REQUEST_PERIOD = time.Millisecond * 500

const (
	TRACKER_START  = 0
	TRACKER_FINISH = 1
)

type TrackerControlMessage struct {
	MessageType int
}

type TrackerStats struct {
	TotalLatency time.Duration
	NRequests    int
}

type requestMessage struct {
	expNum  int64
	latency time.Duration
}

type Tracker struct {
	entrypoint     string
	reqMsgChan     chan requestMessage
	controlInChan  chan TrackerControlMessage
	controlOutChan chan TrackerStats
	isTracking     bool
	expNum         *atomic.Int64
	stats          TrackerStats
}

func CreateTracker(entrypoint string) *Tracker {
	t := Tracker{
		entrypoint:     entrypoint,
		isTracking:     false,
		expNum:         atomic.NewInt64(0),
		reqMsgChan:     make(chan requestMessage),
		controlInChan:  make(chan TrackerControlMessage),
		controlOutChan: make(chan TrackerStats),
	}
	go t.track()
	go t.startWorkload()
	return &t
}

func (t *Tracker) track() {
	fmt.Println("Tracker started!")
	for {
		select {
		case controlMsg := <-t.controlInChan:
			if controlMsg.MessageType == TRACKER_START {
				t.expNum.Inc()
				t.isTracking = true
				t.stats = TrackerStats{}
				t.controlOutChan <- TrackerStats{}
				fmt.Println("Tracker: started tracking")
			} else {
				t.isTracking = false
				t.controlOutChan <- t.stats
				fmt.Println("Tracker: finished tracking")
			}
		case reqMsg := <-t.reqMsgChan:
			if !t.isTracking || reqMsg.expNum != t.expNum.Load() {
				// Ignore request responses if not currently tracking
				// or if request originated during a previous experiment
				continue
			}
			t.stats.NRequests += 1
			t.stats.TotalLatency += reqMsg.latency
		}
	}
}

func (t *Tracker) startWorkload() {
	fmt.Println("Tracker workload started!")
	reqTicker := time.NewTicker(REQUEST_PERIOD)
	for {
		<-reqTicker.C
		go t.makeGetRequest()
	}
}

func (t *Tracker) makeGetRequest() {
	expNum := t.expNum.Load()
	timeBefore := time.Now()
	http.Get(t.entrypoint)
	latency := time.Since(timeBefore)
	reqMsg := requestMessage{
		expNum:  expNum,
		latency: latency,
	}
	t.reqMsgChan <- reqMsg
}

func (t *Tracker) StartTracking() {
	t.controlInChan <- TrackerControlMessage{MessageType: TRACKER_START}
	<-t.controlOutChan
}

func (t *Tracker) FinishTracking() TrackerStats {
	t.controlInChan <- TrackerControlMessage{MessageType: TRACKER_FINISH}
	stats := <-t.controlOutChan
	return stats
}
