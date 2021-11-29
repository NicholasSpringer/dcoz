package client

import (
	"net"

	"github.com/dcoz-controller/utils"
)

// for testing purposes
type UDPClient struct {
	port   int
	socket net.Conn
}

// create a message --> should contain process id to virtually speed things up 

func CreateClient() (client *UDPClient, err error) {
	addr, err := net.ResolveUDPAddr("udp4", ":"+string(utils.PORT))
	if err != nil {
		return nil, err
	}
	socket, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		return nil, err
	}
	return &UDPClient{
		port:   utils.PORT,
		socket: socket,
	}, nil
}

func (c *UDPClient) Accept() {
	for {

	}
}

func (c *UDPClient) DestroyClient() {
	c.socket.Close()
}
