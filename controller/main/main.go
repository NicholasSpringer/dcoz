package main

import (
	"github.com/abiosoft/ishell"
	"github.com/dcoz-controller/experiment"
	"github.com/dcoz-controller/server"
	"github.com/dcoz-controller/utils"
)

func createShell(s *server.UDPServer) {
	shell := ishell.New()
	shell.Println("dcoz")
	shell.AddCmd(&ishell.Cmd{
		Name: "run",
		Help: "runs a set of experiments, returns the results in a txt file",
		Func: func(c *ishell.Context) {
			workload, err := utils.ReadCSVData(c.Args[0])
			if err != nil {
				c.Println("could not read csv")
				return
			}
			entryPoint := c.Args[1]
			containers := c.Args[2:]
			err = experiment.RunExperiments(s, workload, containers, entryPoint)
			if err != nil {
				c.Println("coould not run experiments")
			}
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
