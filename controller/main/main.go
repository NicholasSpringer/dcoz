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
	shell.Run()
}

func main() {

	udpServer, err := server.CreateServer()
	if err != nil {
		// error
		return
	}
	createShell(udpServer)
}
