package main

import (
	"github.com/abiosoft/ishell"
	"github.com/dcoz-controller/server"
)

func createShell(server *server.UDPServer) {
	shell := ishell.New()
	shell.Println("dcoz")
	shell.AddCmd(&ishell.Cmd{
		Name: "run",
		Help: "Runs an experiment",
		Func: func(c *ishell.Context) {
			c.Println("test")
		},
	})
	// running shell=
	shell.Run()
}

func main() {

	// find microservices in soa
	// save values in map
	// run experiment
	// 		--> go through list of services associated with request
	// 		--> send slowdown messages to all other services
	// 		--> pass requests to service from workload generator (need an api for that)
	// 		--> once experiment over, send a message to the servers to reset the scheduling policy

	udpServer, err := server.CreateServer()
	if err != nil {
		// error
		return
	}
	createShell(udpServer)
}
