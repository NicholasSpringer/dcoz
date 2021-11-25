package main

import "github.com/dcoz-controller/server"

func main() {
	udpServer, err := server.CreateServer()
	if err != nil {
		// error
		return
	}
	go udpServer.Serve()

}
