package server

import "net"

var PORT int = 4444
var BUFFSIZE = 1024

type UDPServer struct {
	port int
	conn net.Conn
}

func CreateServer() (server *UDPServer, err error) {
	addr, err := net.ResolveUDPAddr("udp4", ":"+string(PORT))
	// check to see if connection is created
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return nil, err
	}
	return &UDPServer{
		port: PORT,
		conn: conn,
	}, nil
}

func (s *UDPServer) Serve() {
	// buff := make([]byte, BUFFSIZE)
	for {
		// n, addr, err := s.conn.ReadFromUDP(buff)
		// data
	}
}

func (s *UDPServer) DestroyServer() {
	s.conn.Close()
}
