package server

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/dcoz-controller/utils"
)

type PIDMessage struct {
	ProcessID int `json:"processId"`
}

type Message struct {
	SlowdownSpeed int `json:slowdownSpeed`
}

type UDPServer struct {
	port int
	// map of process ids to ip addresses
	processIds map[int]*net.UDPAddr
	exit       chan bool
}

func CreateServer() (server *UDPServer, err error) {
	exitChan := make(chan bool, 1)
	s := &UDPServer{
		port: utils.PORT,
		// for gracefully ending the server
		exit: exitChan,
	}

	// set up listening port first
	go s.listen()
	return s, nil
}

// listens for initial message from agent with process ID --> still need to figure out when to stop listening for init messages
func (s *UDPServer) listen() {
	addr, err := net.ResolveUDPAddr("udp", string(s.port))
	if err != nil {
		fmt.Errorf("received err: %v, expected nil", err)
	}
	conn, err := net.ListenUDP("udp", addr)
	defer conn.Close()
	var message PIDMessage
	for {
		// if destroy server is called, ends the listening routine
		select {
		case <-s.exit:
			return
		default:
		}
		buff := make([]byte, utils.BUFFSIZE)
		// save remote addr, will use to create map for processes
		length, remoteAddr, _ := conn.ReadFromUDP(buff)
		err := json.Unmarshal(buff[:length], &message)
		if err == nil {
			continue
		}
		// adding process ID with addr
		s.processIds[message.ProcessID] = remoteAddr

	}

}

// don't send message to excluded value
func (s *UDPServer) BroadcastSpeedMsg(speed int, excluded int) int {
	numSuccessful := 0
	for pid, addr := range s.processIds {
		if pid == excluded {
			continue
		}
		go func(addr *net.UDPAddr) {
			destAddr, err := net.ResolveUDPAddr("udp", addr.String()+string(s.port))
			if err != nil {
				fmt.Printf("received Error: %v when resolving destination addr, expected nil", err)
				return
			}
			conn, err := net.DialUDP("udp", nil, destAddr)
			// don't keep long lived connection unless necessary
			defer conn.Close()
			if err != nil {
				fmt.Printf("received Error: %v when connecting to destination addr, expected nil", err)
				return
			}
			msg := &Message{SlowdownSpeed: speed}
			encodedMsg, err := json.Marshal(msg)
			if err != nil {
				fmt.Printf("received Error: %v when marshalling data, expected nil", err)
				return
			}
			_, err = conn.Write(encodedMsg)
			if err != nil {
				fmt.Printf("received Error: %v when writing packet, expected nil", err)
				return
			}
			numSuccessful++
		}(addr)
	}
	return numSuccessful
}
