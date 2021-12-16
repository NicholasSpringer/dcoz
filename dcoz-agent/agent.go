package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/NicholasSpringer/dcoz/shared"
)

type pauseConfig struct {
	nCores       int
	prio         int
	pauseBinPath string
}

type contrConn struct {
	conn         *net.TCPConn
	contrMsgChan chan *shared.DcozMessage
	closeChan    chan struct{}
}

type agent struct {
	pauseCfg       pauseConfig
	contr          *contrConn
	newConnChan    chan *contrConn
	updatePidsChan chan []int
	targetPids     []int
	pauseDuration  time.Duration
	pausePeriod    time.Duration
}

func main() {
	var nCores int
	flag.IntVar(&nCores, "cores", 0, "Number of cores in the system")

	var prio int
	flag.IntVar(&prio, "priority", -1, "Priority of blocker threads")

	var pauseBinPath string
	flag.StringVar(&pauseBinPath, "pause", "", "Pausing binary path")

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
	if len(errors) != 0 {
		fmt.Println(strings.Join(errors, "\n"))
		os.Exit(1)
	}
	ag := agent{
		pauseCfg: pauseConfig{
			nCores:       nCores,
			prio:         prio,
			pauseBinPath: pauseBinPath,
		},
		newConnChan:    make(chan *contrConn),
		updatePidsChan: make(chan []int),
	}

	go listenForConnections(ag.newConnChan)
	ag.run()
}

func listenForConnections(newConnChan chan *contrConn) {
	laddr := net.TCPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: shared.AGENT_PORT,
	}
	listener, err := net.ListenTCP("tcp", &laddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while creating TCP listener: %s\n", err)
	}
	fmt.Println("Listening for connections!")
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while waiting for TCP connection: %s\n", err)
			continue
		}
		newContrConn := contrConn{
			conn:         conn,
			contrMsgChan: make(chan *shared.DcozMessage),
			closeChan:    make(chan struct{}),
		}
		fmt.Printf("Listener: New connection from %s\n", newContrConn.conn.RemoteAddr())
		go newContrConn.receiveMessages()
		newConnChan <- &newContrConn
	}
}

func (contr *contrConn) receiveMessages() {
	contr.resetHbDeadline()
	scanner := bufio.NewScanner(contr.conn)
	for {
		if !scanner.Scan() {
			err := scanner.Err()
			contr.conn.Close()
			close(contr.contrMsgChan)
			if err == nil {
				fmt.Printf("Controller %s closed the connection\n", contr.conn.RemoteAddr())
			} else {
				fmt.Fprintf(os.Stderr, "Disconnected from controller %s due to error: %s\n", contr.conn.RemoteAddr(), err)
			}
			return
		}
		contr.resetHbDeadline()
		msg := shared.DcozMessage{}
		shared.UnmarshalDcozMessage([]byte(scanner.Text()), &msg)
		select {
		case _, open := <-contr.closeChan:
			if open {
				panic("No messages should be sent on controller connection close channel!\n")
			}
			fmt.Printf("Closing connection to controller %s due to new connection\n", contr.conn.RemoteAddr())
			contr.conn.Close()
			return
		case contr.contrMsgChan <- &msg:
		}
	}
}

func (contr *contrConn) resetHbDeadline() {
	err := contr.conn.SetDeadline(time.Now().Add(shared.HEARTBEAT_TIMEOUT))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting heartbeat deadline: %s\n", err)
	}
}

func (contr *contrConn) sendResponse(resp byte) {
	respBytes := []byte{resp}
	_, err := contr.conn.Write(respBytes)
	if err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "Error responding to controller %s: %s\n", contr.conn.RemoteAddr(), err)
	}
}

// Listen for messages from controller
func (ag *agent) run() {
	println("Agent started!")
	pauseTicker := time.NewTicker(time.Hour)
	pauseTicker.Stop()
	tickerRunning := false
	for {
		if ag.contr == nil {
			ag.contr = <-ag.newConnChan
			fmt.Printf("Agent: accepted initial controller connection from: %s\n", ag.contr.conn.RemoteAddr())
		}
		select {
		case newContr := <-ag.newConnChan:
			fmt.Printf("Agent: accepting new controller connection from: %s\n", newContr.conn.RemoteAddr())
			close(ag.contr.closeChan)
			if tickerRunning {
				pauseTicker.Stop()
				tickerRunning = false
				ag.targetPids = nil
			}
			ag.contr = newContr
		case msg, open := <-ag.contr.contrMsgChan:
			if !open {
				pauseTicker.Stop()
				tickerRunning = false
				ag.contr = nil
				ag.targetPids = nil
				break
			}
			if msg.MessageType == shared.MSG_UPDATE_TARGETS {
				go updatePids(msg.ContainerIds, ag.updatePidsChan)
				ag.pauseDuration = msg.PauseDuration
				if !tickerRunning || ag.pausePeriod != msg.PausePeriod {
					fmt.Printf("Agent: starting new timer with period: %d ms\n", msg.PausePeriod.Milliseconds())
					ag.pausePeriod = msg.PausePeriod
					if msg.PausePeriod == 0 || msg.PauseDuration == 0 {
						if tickerRunning {
							pauseTicker.Stop()
							tickerRunning = false
							ag.targetPids = nil
						}
					} else {
						if tickerRunning {
							pauseTicker.Reset(msg.PausePeriod)
						} else {
							pauseTicker = time.NewTicker(msg.PausePeriod)
							tickerRunning = true
						}
					}
				}
				if msg.SendResponse {
					fmt.Println("Agent: Synchronizing on new targets")
					newPids := <-ag.updatePidsChan
					ag.targetPids = newPids
					ag.contr.sendResponse(shared.AGENT_RESPONSE_SUCCESS)
				}
			}
		case newPids := <-ag.updatePidsChan:
			fmt.Printf("Agent: Updating target pids to %d entries\n", len(newPids))
			ag.targetPids = newPids
		case <-pauseTicker.C:
			if ag.targetPids != nil && ag.pauseDuration != 0 {
				pause(&ag.pauseCfg, ag.pauseDuration, ag.targetPids)
			}
		}
	}
}

func updatePids(targetContainers []string, updatePidsChan chan []int) {
	updatePidsChan <- getTargetPids(targetContainers)
}
