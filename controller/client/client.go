package client

import (
	"net"

	"github.com/dcoz-controller/utils"
)

// for testing purposes
type UDPClient struct {
	port int
	conn net.Conn
}

func CreateClient() (client *UDPClient, err error) {
	addr, err := net.ResolveUDPAddr("udp4", ":"+string(utils.PORT))
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		return nil, err
	}
	return &UDPClient{
		port: utils.PORT,
		conn: conn,
	}, nil
}

func (c *UDPClient) Accept() {
	for {

	}
}
