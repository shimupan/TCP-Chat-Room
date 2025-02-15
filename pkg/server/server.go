package server

import (
	"fmt"
	"net"
)

type Server struct {
	addr     string
	listener net.Listener
	quitch   chan struct{}
}

func NewServer(addr string) *Server {
	return &Server{
		addr:   addr,
		quitch: make(chan struct{}),
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer listener.Close()
	s.listener = listener
	fmt.Printf("Server started listening on: %s\n", s.listener.Addr().String())

	go s.acceptConn()

	<-s.quitch

	return nil
}

func (s *Server) acceptConn() {
	fmt.Printf("Server accepting connections\n")
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Printf("Error accept: %s\n", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	fmt.Printf("New connection from %s\n", conn.RemoteAddr())
}
